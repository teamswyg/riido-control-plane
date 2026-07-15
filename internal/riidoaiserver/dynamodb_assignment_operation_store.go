package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) loop() {
	defer close(s.done)
	var cachedCredentials AWSCredentials
	for cmd := range s.commands {
		if cmd.close {
			cmd.reply <- nil
			return
		}
		credentials, err := cachedAWSCredentials(cmd.ctx, s.now, s.credentialsProvider, &cachedCredentials)
		if err != nil {
			switch {
			case cmd.load:
				cmd.loadReply <- dynamoDBAssignmentOperationLoadResult{err: err}
			case cmd.queue:
				cmd.queueReply <- dynamoDBAssignmentQueueResult{err: err}
			case cmd.claim:
				cmd.claimReply <- dynamoDBAssignmentClaimResult{err: err}
			case cmd.active:
				cmd.activeReply <- dynamoDBAssignmentActiveLeaseResult{err: err}
			case cmd.projection:
				cmd.projectionReply <- dynamoDBAssignmentProjectionResult{err: err}
			default:
				cmd.reply <- err
			}
			continue
		}
		if cmd.load {
			records, err := s.load(cmd.ctx, credentials)
			cmd.loadReply <- dynamoDBAssignmentOperationLoadResult{records: records, err: err}
			continue
		}
		if cmd.queue {
			assignments, err := s.loadAgentQueue(cmd.ctx, cmd.agentID, credentials)
			cmd.queueReply <- dynamoDBAssignmentQueueResult{assignments: assignments, err: err}
			continue
		}
		if cmd.claim {
			result, err := s.claimNext(cmd.ctx, cmd.agentID, cmd.claimAt, credentials)
			cmd.claimReply <- dynamoDBAssignmentClaimResult{result: result, err: err}
			continue
		}
		if cmd.active {
			lease, found, err := s.loadAgentActiveAssignment(cmd.ctx, cmd.agentID, credentials)
			cmd.activeReply <- dynamoDBAssignmentActiveLeaseResult{lease: lease, found: found, err: err}
			continue
		}
		if cmd.projection {
			record, found, err := s.loadAssignmentProjection(cmd.ctx, cmd.assignmentID, credentials)
			cmd.projectionReply <- dynamoDBAssignmentProjectionResult{projection: AssignmentProjection(record), found: found, err: err}
			continue
		}
		if cmd.refresh {
			if cmd.assignment == nil {
				cmd.reply <- errors.New("riidoaiserver: nil DynamoDB active assignment refresh")
				continue
			}
			cmd.reply <- s.refreshAgentActiveAssignment(cmd.ctx, *cmd.assignment, cmd.refreshAt, credentials)
			continue
		}
		if cmd.record == nil {
			cmd.reply <- errors.New("riidoaiserver: nil DynamoDB assignment operation")
			continue
		}
		cmd.reply <- s.save(cmd.ctx, *cmd.record, credentials)
	}
}

func (s *DynamoDBAssignmentOperationStore) claimNext(ctx context.Context, agentID string, at time.Time, credentials AWSCredentials) (AssignmentClaimResult, error) {
	queue, err := s.loadAgentQueueRecords(ctx, agentID, credentials)
	if err != nil {
		return AssignmentClaimResult{}, err
	}
	for _, candidate := range queue {
		assignment := candidate.Assignment
		if assignment.State.Code() != AssignmentStateCodeQueued {
			continue
		}
		if assignmentQueuedPastMaxAge(assignment, at) {
			repair := dynamoDBAssignmentClaimRepair{
				Operation:            dynamoDBFailStaleQueuedOperation(candidate, at),
				ExpectedState:        AssignmentQueued,
				ExpectedLastEventSeq: candidate.LastEventSeq,
			}
			payload, err := s.claimRepairOnlyTransactionPayload(repair)
			if err != nil {
				return AssignmentClaimResult{}, err
			}
			_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
				endpoint:     s.endpoint,
				endpointHost: s.endpointHost,
				region:       s.region,
				target:       dynamoDBTransactWriteTarget,
				payload:      payload,
				credentials:  credentials,
				httpClient:   s.httpClient,
				now:          s.now,
			})
			if err == nil || isDynamoDBTransactionContention(err) {
				continue
			}
			return AssignmentClaimResult{}, fmt.Errorf("dynamodb repair stale queued assignment: %w", err)
		}
		var repairs []dynamoDBAssignmentClaimRepair
		var activeCondition *dynamoDBTransactWriteConditionCheckAction
		clearMessage := ""
		clearMetadata := map[string]string(nil)
		if assignment.BlockedByAssignmentID != "" {
			blocker, ok, err := s.loadAssignmentProjection(ctx, assignment.BlockedByAssignmentID, credentials)
			if err != nil {
				return AssignmentClaimResult{}, err
			}
			clearMetadata = map[string]string{"blocked_by_assignment_id": assignment.BlockedByAssignmentID}
			switch {
			case !ok:
				clearMessage = "missing blocker cleared before daemon lease"
			case isTerminal(blocker.Assignment.State):
				clearMessage = "terminal blocker cleared before daemon lease"
			case blocker.Assignment.State.Code() == AssignmentStateCodeQueued:
				repairs = append(repairs, dynamoDBAssignmentClaimRepair{
					Operation:            dynamoDBCancelQueuedBlockerOperation(blocker, assignment.ID, at),
					ExpectedState:        AssignmentQueued,
					ExpectedLastEventSeq: blocker.LastEventSeq,
				})
				clearMessage = "queued blocker cleared before daemon lease"
			case assignmentHoldsActiveLease(blocker.Assignment.State):
				stale, condition, err := s.staleBlockerLeaseCondition(ctx, blocker.Assignment, at, credentials)
				if err != nil {
					return AssignmentClaimResult{}, err
				}
				if !stale {
					continue
				}
				activeCondition = condition
				repairs = append(repairs, dynamoDBAssignmentClaimRepair{
					Operation:            dynamoDBFailStaleBlockerOperation(blocker, assignment.ID, at),
					ExpectedState:        blocker.Assignment.State,
					ExpectedLastEventSeq: blocker.LastEventSeq,
				})
				clearMessage = "stale blocker cleared before daemon lease"
			default:
				continue
			}
		}
		operation := dynamoDBClaimAssignmentOperation(assignment, candidate.LastEventSeq, at, clearMessage, clearMetadata)
		payload, err := s.claimTransactionPayload(operation, candidate.LastEventSeq)
		if len(repairs) > 0 || activeCondition != nil {
			payload, err = s.claimRepairTransactionPayload(operation, candidate.LastEventSeq, repairs, activeCondition)
		}
		if err != nil {
			return AssignmentClaimResult{}, err
		}
		_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
			endpoint:     s.endpoint,
			endpointHost: s.endpointHost,
			region:       s.region,
			target:       dynamoDBTransactWriteTarget,
			payload:      payload,
			credentials:  credentials,
			httpClient:   s.httpClient,
			now:          s.now,
		})
		if err == nil {
			operations := make([]AssignmentOperationRecord, 0, len(repairs)+1)
			for _, repair := range repairs {
				operations = append(operations, repair.Operation)
			}
			operations = append(operations, operation)
			return AssignmentClaimResult{Claimed: true, Assignment: operation.Assignment, Operation: operation, Operations: operations}, nil
		}
		if isDynamoDBTransactionContention(err) {
			continue
		}
		return AssignmentClaimResult{}, fmt.Errorf("dynamodb claim assignment: %w", err)
	}
	return AssignmentClaimResult{}, nil
}

type dynamoDBAssignmentClaimRepair struct {
	Operation            AssignmentOperationRecord
	ExpectedState        AssignmentState
	ExpectedLastEventSeq int64
}

func dynamoDBClaimAssignmentOperation(assignment Assignment, lastEventSeq int64, at time.Time, clearMessage string, clearMetadata map[string]string) AssignmentOperationRecord {
	claimed := assignment
	events := []TaskEvent{}
	nextSeq := lastEventSeq
	blockedByID := strings.TrimSpace(claimed.BlockedByAssignmentID)
	if strings.TrimSpace(clearMessage) != "" && blockedByID != "" {
		nextSeq++
		metadata := map[string]string{"blocked_by_assignment_id": blockedByID}
		for key, value := range clearMetadata {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key != "" && value != "" {
				metadata[key] = value
			}
		}
		events = append(events, TaskEvent{
			Seq:          nextSeq,
			TaskID:       claimed.TaskID,
			AssignmentID: claimed.ID,
			AgentID:      claimed.AgentID,
			Type:         EventAssignmentQueued,
			State:        AssignmentQueued,
			Message:      clearMessage,
			Metadata:     metadata,
			At:           at,
		})
		claimed.BlockedByAssignmentID = ""
		claimed.UpdatedAt = at
	}
	claimed.State = AssignmentLeased
	claimed.LeaseToken = fmt.Sprintf("%s:%d", claimed.ID, at.UnixNano())
	claimed.UpdatedAt = at
	nextSeq++
	events = append(events, TaskEvent{
		Seq:          nextSeq,
		TaskID:       claimed.TaskID,
		AssignmentID: claimed.ID,
		AgentID:      claimed.AgentID,
		Type:         EventAssignmentLeased,
		State:        AssignmentLeased,
		Metadata:     map[string]string{"lease_token": claimed.LeaseToken},
		At:           at,
	})
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationPollStart, claimed, events),
		OperationType: AssignmentOperationPollStart,
		TaskID:        claimed.TaskID,
		AssignmentID:  claimed.ID,
		AgentID:       claimed.AgentID,
		Assignment:    claimed,
		Events:        events,
		RecordedAt:    at,
	}
}

func dynamoDBCancelQueuedBlockerOperation(blocker assignmentProjectionRecord, blockedAssignmentID string, at time.Time) AssignmentOperationRecord {
	assignment := blocker.Assignment
	assignment.State = AssignmentCancelled
	assignment.UpdatedAt = at
	events := []TaskEvent{{
		Seq:          blocker.LastEventSeq + 1,
		TaskID:       assignment.TaskID,
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentCancelled,
		State:        AssignmentCancelled,
		Message:      "queued blocker was cancelled before queued assignment claim",
		Metadata:     map[string]string{"blocked_assignment_id": blockedAssignmentID},
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAgentEvent, assignment, events),
		OperationType: AssignmentOperationAgentEvent,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}

func dynamoDBFailStaleBlockerOperation(blocker assignmentProjectionRecord, blockedAssignmentID string, at time.Time) AssignmentOperationRecord {
	assignment := blocker.Assignment
	assignment.State = AssignmentFailed
	assignment.UpdatedAt = at
	metadata := map[string]string{"blocked_assignment_id": blockedAssignmentID}
	if assignment.LeaseToken != "" {
		metadata["lease_token"] = assignment.LeaseToken
	}
	events := []TaskEvent{{
		Seq:          blocker.LastEventSeq + 1,
		TaskID:       assignment.TaskID,
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentFailed,
		State:        AssignmentFailed,
		Message:      "blocked queued assignment repaired after stale blocker lease expired",
		Metadata:     metadata,
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAgentEvent, assignment, events),
		OperationType: AssignmentOperationAgentEvent,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}

func (s *DynamoDBAssignmentOperationStore) staleBlockerLeaseCondition(ctx context.Context, blocker Assignment, at time.Time, credentials AWSCredentials) (bool, *dynamoDBTransactWriteConditionCheckAction, error) {
	if !assignmentHoldsActiveLease(blocker.State) {
		return false, nil, nil
	}
	key := map[string]map[string]string{
		"pk": {"S": dynamoDBAgentActivePK(blocker.AgentID)},
		"sk": {"S": dynamoDBAgentActiveSK},
	}
	lease, exists, err := s.loadAgentActiveAssignment(ctx, blocker.AgentID, credentials)
	if err != nil {
		return false, nil, err
	}
	if !exists {
		return true, &dynamoDBTransactWriteConditionCheckAction{
			TableName:           s.tableName,
			Key:                 key,
			ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
		}, nil
	}
	if lease.ActiveAssignmentID != blocker.ID {
		return true, &dynamoDBTransactWriteConditionCheckAction{
			TableName:           s.tableName,
			Key:                 key,
			ConditionExpression: "active_assignment_id <> :blocked_assignment_id",
			ExpressionAttributeValues: map[string]map[string]string{
				":blocked_assignment_id": {"S": blocker.ID},
			},
		}, nil
	}
	if !lease.Expired(at) {
		return false, nil, nil
	}
	return true, &dynamoDBTransactWriteConditionCheckAction{
		TableName: s.tableName,
		Key:       key,
		ConditionExpression: "active_assignment_id = :blocked_assignment_id AND " +
			"((attribute_exists(lease_expires_unix_ms) AND lease_expires_unix_ms <= :claim_started_unix_ms) OR " +
			"(attribute_not_exists(lease_expires_unix_ms) AND lease_expires_at <= :claim_started_at))",
		ExpressionAttributeValues: map[string]map[string]string{
			":blocked_assignment_id": {"S": blocker.ID},
			":claim_started_unix_ms": {"N": strconv.FormatInt(at.UTC().UnixMilli(), 10)},
			":claim_started_at":      {"S": at.UTC().Format(time.RFC3339Nano)},
		},
	}, nil
}

func (s *DynamoDBAssignmentOperationStore) loadAgentQueue(ctx context.Context, agentID string, credentials AWSCredentials) ([]Assignment, error) {
	records, err := s.loadAgentQueueRecords(ctx, agentID, credentials)
	if err != nil {
		return nil, err
	}
	assignments := make([]Assignment, 0, len(records))
	for _, record := range records {
		assignments = append(assignments, record.Assignment)
	}
	return assignments, nil
}

func (s *DynamoDBAssignmentOperationStore) loadAgentQueueRecords(ctx context.Context, agentID string, credentials AWSCredentials) ([]assignmentProjectionRecord, error) {
	var records []assignmentProjectionRecord
	var startKey map[string]map[string]string
	for {
		payload, err := s.agentQueueQueryPayload(agentID, startKey)
		if err != nil {
			return nil, err
		}
		body, err := doDynamoDBJSON(ctx, dynamoDBRequest{
			endpoint:     s.endpoint,
			endpointHost: s.endpointHost,
			region:       s.region,
			target:       dynamoDBQueryTarget,
			payload:      payload,
			credentials:  credentials,
			httpClient:   s.httpClient,
			now:          s.now,
		})
		if err != nil {
			return nil, fmt.Errorf("dynamodb query assignment queue: %w", err)
		}
		var response struct {
			Items            []map[string]map[string]string `json:"Items"`
			LastEvaluatedKey map[string]map[string]string   `json:"LastEvaluatedKey"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("decode DynamoDB assignment queue response: %w", err)
		}
		for _, item := range response.Items {
			record, err := assignmentProjectionFromDynamoDBItem(item)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}
		if len(response.LastEvaluatedKey) == 0 {
			return records, nil
		}
		startKey = response.LastEvaluatedKey
	}
}

func (s *DynamoDBAssignmentOperationStore) loadAssignmentProjection(ctx context.Context, assignmentID string, credentials AWSCredentials) (assignmentProjectionRecord, bool, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key: map[string]map[string]string{
			"pk": {"S": dynamoDBAssignmentProjectionPK(assignmentID)},
			"sk": {"S": dynamoDBAssignmentProjectionSK},
		},
	})
	if err != nil {
		return assignmentProjectionRecord{}, false, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBGetItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err != nil {
		return assignmentProjectionRecord{}, false, fmt.Errorf("dynamodb get assignment projection: %w", err)
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return assignmentProjectionRecord{}, false, fmt.Errorf("decode DynamoDB assignment projection response: %w", err)
	}
	if len(response.Item) == 0 {
		return assignmentProjectionRecord{}, false, nil
	}
	record, err := assignmentProjectionFromDynamoDBItem(response.Item)
	if err != nil {
		return assignmentProjectionRecord{}, false, err
	}
	return record, true, nil
}

func (s *DynamoDBAssignmentOperationStore) loadAgentActiveAssignment(ctx context.Context, agentID string, credentials AWSCredentials) (AssignmentActiveLease, bool, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key: map[string]map[string]string{
			"pk": {"S": dynamoDBAgentActivePK(agentID)},
			"sk": {"S": dynamoDBAgentActiveSK},
		},
	})
	if err != nil {
		return AssignmentActiveLease{}, false, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBGetItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err != nil {
		return AssignmentActiveLease{}, false, fmt.Errorf("dynamodb get agent active assignment: %w", err)
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return AssignmentActiveLease{}, false, fmt.Errorf("decode DynamoDB agent active assignment response: %w", err)
	}
	if len(response.Item) == 0 {
		return AssignmentActiveLease{}, false, nil
	}
	lease, err := agentActiveAssignmentFromDynamoDBItem(response.Item)
	if err != nil {
		return AssignmentActiveLease{}, false, err
	}
	return lease, true, nil
}

func (s *DynamoDBAssignmentOperationStore) save(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	if record.SchemaVersion == "" {
		record.SchemaVersion = AssignmentOperationSchemaVersion
	}
	if record.RecordedAt.IsZero() {
		record.RecordedAt = s.now()
	}
	if err := validateAssignmentOperationRecord(record); err != nil {
		return err
	}
	payload, err := s.putOperationPayload(record)
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBPutItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err == nil {
		return s.saveAssignmentProjection(ctx, record, credentials)
	}
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("dynamodb put assignment operation: %w", err)
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	if strings.Contains(dynamoErr.Type, "ConditionalCheckFailedException") {
		return s.saveAssignmentProjection(ctx, record, credentials)
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = dynamoErr.MessageAlt
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = strings.TrimSpace(string(apiErr.body))
	}
	return fmt.Errorf("dynamodb put assignment operation: status=%d type=%q message=%q", apiErr.statusCode, dynamoErr.Type, dynamoErr.Message)
}

func (s *DynamoDBAssignmentOperationStore) saveAssignmentProjection(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	payload, err := s.putAssignmentProjectionPayload(record)
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBPutItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err == nil {
		return s.releaseAgentActiveAssignmentIfTerminal(ctx, record, credentials)
	}
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("dynamodb put assignment projection: %w", err)
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	if strings.Contains(dynamoErr.Type, "ConditionalCheckFailedException") {
		return s.releaseAgentActiveAssignmentIfTerminal(ctx, record, credentials)
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = dynamoErr.MessageAlt
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = strings.TrimSpace(string(apiErr.body))
	}
	return fmt.Errorf("dynamodb put assignment projection: status=%d type=%q message=%q", apiErr.statusCode, dynamoErr.Type, dynamoErr.Message)
}

func (s *DynamoDBAssignmentOperationStore) releaseAgentActiveAssignmentIfTerminal(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	if !isTerminal(record.Assignment.State) {
		return nil
	}
	payload, err := s.deleteAgentActiveAssignmentPayload(record)
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBDeleteItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err == nil {
		return nil
	}
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("dynamodb delete agent active assignment: %w", err)
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	if strings.Contains(dynamoErr.Type, "ConditionalCheckFailedException") {
		return nil
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = dynamoErr.MessageAlt
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = strings.TrimSpace(string(apiErr.body))
	}
	return fmt.Errorf("dynamodb delete agent active assignment: status=%d type=%q message=%q", apiErr.statusCode, dynamoErr.Type, dynamoErr.Message)
}

func (s *DynamoDBAssignmentOperationStore) refreshAgentActiveAssignment(ctx context.Context, assignment Assignment, at time.Time, credentials AWSCredentials) error {
	if !assignmentHoldsActiveLease(assignment.State) {
		return nil
	}
	if strings.TrimSpace(assignment.AgentID) == "" || strings.TrimSpace(assignment.ID) == "" {
		return errors.New("riidoaiserver: active assignment refresh requires agent_id and assignment_id")
	}
	if strings.TrimSpace(assignment.LeaseToken) == "" {
		return errors.New("riidoaiserver: active assignment refresh requires lease_token")
	}
	payload, err := s.updateAgentActiveAssignmentPayload(assignment, at)
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBUpdateItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err == nil {
		return nil
	}
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("dynamodb update agent active assignment: %w", err)
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	if strings.Contains(dynamoErr.Type, "ConditionalCheckFailedException") {
		return nil
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = dynamoErr.MessageAlt
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = strings.TrimSpace(string(apiErr.body))
	}
	return fmt.Errorf("dynamodb update agent active assignment: status=%d type=%q message=%q", apiErr.statusCode, dynamoErr.Type, dynamoErr.Message)
}

func (s *DynamoDBAssignmentOperationStore) claimTransactionPayload(record AssignmentOperationRecord, expectedLastEventSeq int64) ([]byte, error) {
	operationItem, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	projectionItem, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	payload := struct {
		TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
	}{
		TransactItems: make([]dynamoDBTransactWritePut, 3),
	}
	payload.TransactItems[0].Put.TableName = s.tableName
	payload.TransactItems[0].Put.ConditionExpression = "assignment_state = :queued AND agent_id = :agent_id AND last_event_seq = :expected_last_event_seq"
	payload.TransactItems[0].Put.ExpressionAttributeValues = map[string]map[string]string{
		":queued":                  {"S": string(AssignmentQueued)},
		":agent_id":                {"S": record.AgentID},
		":expected_last_event_seq": {"N": strconv.FormatInt(expectedLastEventSeq, 10)},
	}
	payload.TransactItems[0].Put.Item = projectionItem
	payload.TransactItems[1].Put.TableName = s.tableName
	payload.TransactItems[1].Put.ConditionExpression = "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR lease_expires_unix_ms <= :claim_started_unix_ms"
	payload.TransactItems[1].Put.ExpressionAttributeValues = map[string]map[string]string{
		":claim_started_unix_ms": {"N": strconv.FormatInt(record.RecordedAt.UTC().UnixMilli(), 10)},
	}
	payload.TransactItems[1].Put.Item = s.agentActiveAssignmentDynamoDBItem(record)
	payload.TransactItems[2].Put.TableName = s.tableName
	payload.TransactItems[2].Put.ConditionExpression = "attribute_not_exists(pk) AND attribute_not_exists(sk)"
	payload.TransactItems[2].Put.Item = operationItem
	return json.Marshal(payload)
}

func (s *DynamoDBAssignmentOperationStore) claimRepairTransactionPayload(record AssignmentOperationRecord, expectedLastEventSeq int64, repairs []dynamoDBAssignmentClaimRepair, activeCondition *dynamoDBTransactWriteConditionCheckAction) ([]byte, error) {
	items := []dynamoDBTransactWriteItem{}
	if activeCondition != nil {
		items = append(items, dynamoDBTransactWriteItem{ConditionCheck: activeCondition})
	}
	for _, repair := range repairs {
		projectionItem, err := assignmentProjectionDynamoDBItem(repair.Operation)
		if err != nil {
			return nil, err
		}
		operationItem, err := assignmentOperationDynamoDBItem(repair.Operation)
		if err != nil {
			return nil, err
		}
		items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
			TableName:           s.tableName,
			ConditionExpression: "assignment_state = :expected_state AND last_event_seq = :expected_last_event_seq",
			ExpressionAttributeValues: map[string]map[string]string{
				":expected_state":          {"S": string(repair.ExpectedState)},
				":expected_last_event_seq": {"N": strconv.FormatInt(repair.ExpectedLastEventSeq, 10)},
			},
			Item: projectionItem,
		}})
		items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
			TableName:           s.tableName,
			ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
			Item:                operationItem,
		}})
	}
	projectionItem, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	operationItem, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
		TableName:           s.tableName,
		ConditionExpression: "assignment_state = :queued AND agent_id = :agent_id AND last_event_seq = :expected_last_event_seq",
		ExpressionAttributeValues: map[string]map[string]string{
			":queued":                  {"S": string(AssignmentQueued)},
			":agent_id":                {"S": record.AgentID},
			":expected_last_event_seq": {"N": strconv.FormatInt(expectedLastEventSeq, 10)},
		},
		Item: projectionItem,
	}})
	items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
		TableName:           s.tableName,
		ConditionExpression: "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR lease_expires_unix_ms <= :claim_started_unix_ms",
		ExpressionAttributeValues: map[string]map[string]string{
			":claim_started_unix_ms": {"N": strconv.FormatInt(record.RecordedAt.UTC().UnixMilli(), 10)},
		},
		Item: s.agentActiveAssignmentDynamoDBItem(record),
	}})
	items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
		TableName:           s.tableName,
		ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
		Item:                operationItem,
	}})
	payload := struct {
		TransactItems []dynamoDBTransactWriteItem `json:"TransactItems"`
	}{
		TransactItems: items,
	}
	return json.Marshal(payload)
}

var (
	_ AssignmentOperationStore   = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentOperationLoader  = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentQueueReader      = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentClaimer          = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentActiveLeaseStore = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentProjectionReader = (*DynamoDBAssignmentOperationStore)(nil)
)
