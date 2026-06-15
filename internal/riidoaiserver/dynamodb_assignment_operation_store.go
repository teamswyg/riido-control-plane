package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	dynamoDBAssignmentOperationPK  = "ASSIGNMENT_OPERATION"
	dynamoDBAssignmentProjectionSK = "STATE"
	dynamoDBAgentActiveSK          = "ACTIVE"
	dynamoDBAssignmentQueueIndex   = "agent_queue"
	dynamoDBAssignmentQueryLimit   = 50
)

type DynamoDBAssignmentOperationStoreConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	ActiveLeaseDuration time.Duration
	CredentialsProvider AWSCredentialsProvider
}

type DynamoDBAssignmentOperationStore struct {
	commands            chan dynamoDBAssignmentOperationCommand
	done                chan struct{}
	region              string
	tableName           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	activeLeaseDuration time.Duration
	credentialsProvider AWSCredentialsProvider
}

type dynamoDBAssignmentOperationCommand struct {
	ctx             context.Context
	load            bool
	queue           bool
	claim           bool
	active          bool
	refresh         bool
	projection      bool
	agentID         string
	assignmentID    string
	claimAt         time.Time
	refreshAt       time.Time
	record          *AssignmentOperationRecord
	assignment      *Assignment
	close           bool
	reply           chan error
	loadReply       chan dynamoDBAssignmentOperationLoadResult
	queueReply      chan dynamoDBAssignmentQueueResult
	claimReply      chan dynamoDBAssignmentClaimResult
	activeReply     chan dynamoDBAssignmentActiveLeaseResult
	projectionReply chan dynamoDBAssignmentProjectionResult
}

type dynamoDBAssignmentOperationLoadResult struct {
	records []AssignmentOperationRecord
	err     error
}

type dynamoDBAssignmentQueueResult struct {
	assignments []Assignment
	err         error
}

type dynamoDBAssignmentClaimResult struct {
	result AssignmentClaimResult
	err    error
}

type dynamoDBAssignmentActiveLeaseResult struct {
	lease AssignmentActiveLease
	found bool
	err   error
}

type dynamoDBAssignmentProjectionResult struct {
	projection AssignmentProjection
	found      bool
	err        error
}

type assignmentProjectionRecord struct {
	Assignment   Assignment
	LastEventSeq int64
}

func NewDynamoDBAssignmentOperationStore(config DynamoDBAssignmentOperationStoreConfig) (*DynamoDBAssignmentOperationStore, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store table name is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store credentials provider is required")
	}
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	activeLeaseDuration := config.ActiveLeaseDuration
	if activeLeaseDuration <= 0 {
		activeLeaseDuration = time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second
	}
	store := &DynamoDBAssignmentOperationStore{
		commands:            make(chan dynamoDBAssignmentOperationCommand),
		done:                make(chan struct{}),
		region:              region,
		tableName:           tableName,
		endpoint:            endpoint,
		endpointHost:        endpointHost,
		httpClient:          dynamoDBHTTPClient(config.HTTPClient),
		now:                 dynamoDBClock(config.Now),
		activeLeaseDuration: activeLeaseDuration,
		credentialsProvider: config.CredentialsProvider,
	}
	go store.loop()
	return store, nil
}

func (s *DynamoDBAssignmentOperationStore) SaveAssignmentOperation(ctx context.Context, record AssignmentOperationRecord) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan error, 1)
	recordCopy := record
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{ctx: ctx, record: &recordCopy, reply: reply}:
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) LoadAssignmentOperations(ctx context.Context) ([]AssignmentOperationRecord, error) {
	if s == nil {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan dynamoDBAssignmentOperationLoadResult, 1)
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{ctx: ctx, load: true, loadReply: reply}:
	case <-s.done:
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case result := <-reply:
		return result.records, result.err
	case <-s.done:
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) LoadAgentQueueAssignments(ctx context.Context, agentID string) ([]Assignment, error) {
	if s == nil {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return nil, errors.New("riidoaiserver: assignment queue agent_id is required")
	}
	reply := make(chan dynamoDBAssignmentQueueResult, 1)
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{ctx: ctx, queue: true, agentID: agentID, queueReply: reply}:
	case <-s.done:
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case result := <-reply:
		return result.assignments, result.err
	case <-s.done:
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) ClaimNextAssignment(ctx context.Context, agentID string, at time.Time) (AssignmentClaimResult, error) {
	if s == nil {
		return AssignmentClaimResult{}, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AssignmentClaimResult{}, errors.New("riidoaiserver: assignment claim agent_id is required")
	}
	if at.IsZero() {
		at = s.now()
	}
	reply := make(chan dynamoDBAssignmentClaimResult, 1)
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{ctx: ctx, claim: true, agentID: agentID, claimAt: at, claimReply: reply}:
	case <-s.done:
		return AssignmentClaimResult{}, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentClaimResult{}, ctx.Err()
	}
	select {
	case result := <-reply:
		return result.result, result.err
	case <-s.done:
		return AssignmentClaimResult{}, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentClaimResult{}, ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) LoadAgentActiveAssignment(ctx context.Context, agentID string) (AssignmentActiveLease, bool, error) {
	if s == nil {
		return AssignmentActiveLease{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AssignmentActiveLease{}, false, errors.New("riidoaiserver: active lease agent_id is required")
	}
	reply := make(chan dynamoDBAssignmentActiveLeaseResult, 1)
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{ctx: ctx, active: true, agentID: agentID, activeReply: reply}:
	case <-s.done:
		return AssignmentActiveLease{}, false, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentActiveLease{}, false, ctx.Err()
	}
	select {
	case result := <-reply:
		return result.lease, result.found, result.err
	case <-s.done:
		return AssignmentActiveLease{}, false, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentActiveLease{}, false, ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) LoadAssignmentProjection(ctx context.Context, assignmentID string) (AssignmentProjection, bool, error) {
	if s == nil {
		return AssignmentProjection{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AssignmentProjection{}, false, errors.New("riidoaiserver: assignment projection assignment_id is required")
	}
	reply := make(chan dynamoDBAssignmentProjectionResult, 1)
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{ctx: ctx, projection: true, assignmentID: assignmentID, projectionReply: reply}:
	case <-s.done:
		return AssignmentProjection{}, false, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentProjection{}, false, ctx.Err()
	}
	select {
	case result := <-reply:
		return result.projection, result.found, result.err
	case <-s.done:
		return AssignmentProjection{}, false, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentProjection{}, false, ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) RefreshAgentActiveAssignment(ctx context.Context, assignment Assignment, at time.Time) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if at.IsZero() {
		at = s.now()
	}
	assignmentCopy := assignment
	reply := make(chan error, 1)
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{ctx: ctx, refresh: true, assignment: &assignmentCopy, refreshAt: at, reply: reply}:
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) Close() error {
	if s == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{close: true, reply: reply}:
		return <-reply
	case <-s.done:
		return nil
	}
}

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
			if cmd.load {
				cmd.loadReply <- dynamoDBAssignmentOperationLoadResult{err: err}
			} else if cmd.queue {
				cmd.queueReply <- dynamoDBAssignmentQueueResult{err: err}
			} else if cmd.claim {
				cmd.claimReply <- dynamoDBAssignmentClaimResult{err: err}
			} else if cmd.active {
				cmd.activeReply <- dynamoDBAssignmentActiveLeaseResult{err: err}
			} else if cmd.projection {
				cmd.projectionReply <- dynamoDBAssignmentProjectionResult{err: err}
			} else {
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

func (s *DynamoDBAssignmentOperationStore) load(ctx context.Context, credentials AWSCredentials) ([]AssignmentOperationRecord, error) {
	var records []AssignmentOperationRecord
	var startKey map[string]map[string]string
	for {
		payload, err := s.queryPayload(startKey)
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
			return nil, fmt.Errorf("dynamodb query assignment operations: %w", err)
		}
		if len(bytes.TrimSpace(body)) == 0 {
			body = []byte(`{}`)
		}
		var response struct {
			Items            []map[string]map[string]string `json:"Items"`
			LastEvaluatedKey map[string]map[string]string   `json:"LastEvaluatedKey"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("decode DynamoDB assignment operation query response: %w", err)
		}
		for _, item := range response.Items {
			record, err := assignmentOperationRecordFromDynamoDBItem(item)
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

func (s *DynamoDBAssignmentOperationStore) putOperationPayload(record AssignmentOperationRecord) ([]byte, error) {
	item, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	payload := struct {
		TableName           string                       `json:"TableName"`
		ConditionExpression string                       `json:"ConditionExpression"`
		Item                map[string]map[string]string `json:"Item"`
	}{
		TableName:           s.tableName,
		ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
		Item:                item,
	}
	return json.Marshal(payload)
}

func assignmentOperationDynamoDBItem(record AssignmentOperationRecord) (map[string]map[string]string, error) {
	assignmentJSON, err := json.Marshal(record.Assignment)
	if err != nil {
		return nil, err
	}
	eventsJSON, err := json.Marshal(record.Events)
	if err != nil {
		return nil, err
	}
	item := map[string]map[string]string{
		"pk":                 {"S": dynamoDBAssignmentOperationPK},
		"sk":                 {"S": assignmentOperationSortKey(record)},
		"schema_version":     {"S": record.SchemaVersion},
		"operation_id":       {"S": record.OperationID},
		"operation_type":     {"S": string(record.OperationType)},
		"task_id":            {"S": record.TaskID},
		"assignment_id":      {"S": record.AssignmentID},
		"operation_agent_id": {"S": record.AgentID},
		"assignment_state":   {"S": string(record.Assignment.State)},
		"recorded_at":        {"S": record.RecordedAt.UTC().Format(time.RFC3339Nano)},
		"last_event_seq":     {"N": strconv.FormatInt(assignmentOperationLastEventSeq(record), 10)},
		"event_count":        {"N": strconv.Itoa(len(record.Events))},
		"assignment_json":    {"S": string(assignmentJSON)},
		"events_json":        {"S": string(eventsJSON)},
	}
	return item, nil
}

func (s *DynamoDBAssignmentOperationStore) putAssignmentProjectionPayload(record AssignmentOperationRecord) ([]byte, error) {
	item, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	lastEventSeq := assignmentOperationLastEventSeq(record)
	payload := struct {
		TableName                 string                       `json:"TableName"`
		ConditionExpression       string                       `json:"ConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		Item                      map[string]map[string]string `json:"Item"`
	}{
		TableName:           s.tableName,
		ConditionExpression: "attribute_not_exists(last_event_seq) OR last_event_seq <= :last_event_seq",
		ExpressionAttributeValues: map[string]map[string]string{
			":last_event_seq": {"N": strconv.FormatInt(lastEventSeq, 10)},
		},
		Item: item,
	}
	return json.Marshal(payload)
}

func assignmentProjectionDynamoDBItem(record AssignmentOperationRecord) (map[string]map[string]string, error) {
	assignmentJSON, err := json.Marshal(record.Assignment)
	if err != nil {
		return nil, err
	}
	lastEventSeq := assignmentOperationLastEventSeq(record)
	item := map[string]map[string]string{
		"pk":                 {"S": dynamoDBAssignmentProjectionPK(record.AssignmentID)},
		"sk":                 {"S": dynamoDBAssignmentProjectionSK},
		"schema_version":     {"S": AssignmentProjectionSchemaVersion},
		"assignment_id":      {"S": record.AssignmentID},
		"task_id":            {"S": record.TaskID},
		"component_id":       {"S": record.Assignment.ComponentID},
		"operation_id":       {"S": record.OperationID},
		"operation_type":     {"S": string(record.OperationType)},
		"operation_agent_id": {"S": record.AgentID},
		"assignment_state":   {"S": string(record.Assignment.State)},
		"runtime_provider":   {"S": record.Assignment.RuntimeProvider},
		"prompt":             {"S": record.Assignment.Prompt},
		"created_at":         {"S": record.Assignment.CreatedAt.UTC().Format(time.RFC3339Nano)},
		"updated_at":         {"S": record.Assignment.UpdatedAt.UTC().Format(time.RFC3339Nano)},
		"recorded_at":        {"S": record.RecordedAt.UTC().Format(time.RFC3339Nano)},
		"last_event_seq":     {"N": strconv.FormatInt(lastEventSeq, 10)},
		"assignment_json":    {"S": string(assignmentJSON)},
	}
	if record.Assignment.LeaseToken != "" {
		item["lease_token"] = map[string]string{"S": record.Assignment.LeaseToken}
	}
	if record.Assignment.AgentInstruction != "" {
		item["agent_instruction"] = map[string]string{"S": record.Assignment.AgentInstruction}
	}
	if record.Assignment.ReplacesAssignmentID != "" {
		item["replaces_assignment_id"] = map[string]string{"S": record.Assignment.ReplacesAssignmentID}
	}
	if record.Assignment.BlockedByAssignmentID != "" {
		item["blocked_by_assignment_id"] = map[string]string{"S": record.Assignment.BlockedByAssignmentID}
	}
	if record.Assignment.State.Code() == AssignmentStateCodeQueued {
		item["agent_id"] = map[string]string{"S": record.Assignment.AgentID}
		item["assignment_sort"] = map[string]string{"S": assignmentQueueSort(record.Assignment)}
	}
	return item, nil
}

func (s *DynamoDBAssignmentOperationStore) agentActiveAssignmentDynamoDBItem(record AssignmentOperationRecord) map[string]map[string]string {
	leaseHeartbeatAt := record.RecordedAt.UTC()
	leaseExpiresAt := s.activeLeaseExpiresAt(leaseHeartbeatAt)
	item := map[string]map[string]string{
		"pk":                    {"S": dynamoDBAgentActivePK(record.AgentID)},
		"sk":                    {"S": dynamoDBAgentActiveSK},
		"schema_version":        {"S": AssignmentAgentActiveSchemaVersion},
		"agent_id":              {"S": record.AgentID},
		"active_assignment_id":  {"S": record.AssignmentID},
		"task_id":               {"S": record.TaskID},
		"component_id":          {"S": record.Assignment.ComponentID},
		"lease_token":           {"S": record.Assignment.LeaseToken},
		"operation_id":          {"S": record.OperationID},
		"operation_type":        {"S": string(record.OperationType)},
		"runtime_provider":      {"S": record.Assignment.RuntimeProvider},
		"assignment_state":      {"S": string(record.Assignment.State)},
		"leased_at":             {"S": leaseHeartbeatAt.Format(time.RFC3339Nano)},
		"lease_heartbeat_at":    {"S": leaseHeartbeatAt.Format(time.RFC3339Nano)},
		"lease_expires_at":      {"S": leaseExpiresAt.Format(time.RFC3339Nano)},
		"lease_expires_unix_ms": {"N": strconv.FormatInt(leaseExpiresAt.UnixMilli(), 10)},
		"updated_at":            {"S": record.Assignment.UpdatedAt.UTC().Format(time.RFC3339Nano)},
		"last_event_seq":        {"N": strconv.FormatInt(assignmentOperationLastEventSeq(record), 10)},
	}
	return item
}

func (s *DynamoDBAssignmentOperationStore) activeLeaseExpiresAt(at time.Time) time.Time {
	return at.UTC().Add(s.activeLeaseDuration)
}

func (s *DynamoDBAssignmentOperationStore) deleteAgentActiveAssignmentPayload(record AssignmentOperationRecord) ([]byte, error) {
	payload := struct {
		TableName                 string                       `json:"TableName"`
		Key                       map[string]map[string]string `json:"Key"`
		ConditionExpression       string                       `json:"ConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	}{
		TableName: s.tableName,
		Key: map[string]map[string]string{
			"pk": {"S": dynamoDBAgentActivePK(record.AgentID)},
			"sk": {"S": dynamoDBAgentActiveSK},
		},
		ConditionExpression: "active_assignment_id = :assignment_id",
		ExpressionAttributeValues: map[string]map[string]string{
			":assignment_id": {"S": record.AssignmentID},
		},
	}
	return json.Marshal(payload)
}

func (s *DynamoDBAssignmentOperationStore) updateAgentActiveAssignmentPayload(assignment Assignment, heartbeatAt time.Time) ([]byte, error) {
	heartbeatAt = heartbeatAt.UTC()
	leaseExpiresAt := s.activeLeaseExpiresAt(heartbeatAt)
	payload := struct {
		TableName                 string                       `json:"TableName"`
		Key                       map[string]map[string]string `json:"Key"`
		ConditionExpression       string                       `json:"ConditionExpression"`
		UpdateExpression          string                       `json:"UpdateExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	}{
		TableName: s.tableName,
		Key: map[string]map[string]string{
			"pk": {"S": dynamoDBAgentActivePK(assignment.AgentID)},
			"sk": {"S": dynamoDBAgentActiveSK},
		},
		ConditionExpression: "active_assignment_id = :assignment_id AND lease_token = :lease_token",
		UpdateExpression:    "SET assignment_state = :assignment_state, lease_heartbeat_at = :heartbeat_at, lease_expires_at = :lease_expires_at, lease_expires_unix_ms = :lease_expires_unix_ms, updated_at = :heartbeat_at",
		ExpressionAttributeValues: map[string]map[string]string{
			":assignment_id":         {"S": assignment.ID},
			":lease_token":           {"S": assignment.LeaseToken},
			":assignment_state":      {"S": string(assignment.State)},
			":heartbeat_at":          {"S": heartbeatAt.Format(time.RFC3339Nano)},
			":lease_expires_at":      {"S": leaseExpiresAt.Format(time.RFC3339Nano)},
			":lease_expires_unix_ms": {"N": strconv.FormatInt(leaseExpiresAt.UnixMilli(), 10)},
		},
	}
	return json.Marshal(payload)
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

type dynamoDBTransactWritePut struct {
	Put struct {
		TableName                 string                       `json:"TableName"`
		ConditionExpression       string                       `json:"ConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues,omitempty"`
		Item                      map[string]map[string]string `json:"Item"`
	} `json:"Put"`
}

type dynamoDBTransactWriteItem struct {
	Put            *dynamoDBTransactWritePutAction            `json:"Put,omitempty"`
	ConditionCheck *dynamoDBTransactWriteConditionCheckAction `json:"ConditionCheck,omitempty"`
}

type dynamoDBTransactWritePutAction struct {
	TableName                 string                       `json:"TableName"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues,omitempty"`
	Item                      map[string]map[string]string `json:"Item"`
}

type dynamoDBTransactWriteConditionCheckAction struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues,omitempty"`
}

func isDynamoDBTransactionContention(err error) bool {
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Code       string `json:"Code"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	errorText := dynamoErr.Type + " " + dynamoErr.Code + " " + dynamoErr.Message + " " + dynamoErr.MessageAlt + " " + string(apiErr.body)
	return strings.Contains(errorText, "TransactionCanceledException") ||
		strings.Contains(errorText, "ConditionalCheckFailedException")
}

func (s *DynamoDBAssignmentOperationStore) queryPayload(exclusiveStartKey map[string]map[string]string) ([]byte, error) {
	payload := struct {
		TableName                 string                       `json:"TableName"`
		ConsistentRead            bool                         `json:"ConsistentRead"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
		ScanIndexForward          bool                         `json:"ScanIndexForward"`
		Limit                     int                          `json:"Limit"`
	}{
		TableName:              s.tableName,
		ConsistentRead:         true,
		KeyConditionExpression: "pk = :pk",
		ExpressionAttributeValues: map[string]map[string]string{
			":pk": {"S": dynamoDBAssignmentOperationPK},
		},
		ExclusiveStartKey: exclusiveStartKey,
		ScanIndexForward:  true,
		Limit:             dynamoDBAssignmentQueryLimit,
	}
	return json.Marshal(payload)
}

func (s *DynamoDBAssignmentOperationStore) agentQueueQueryPayload(agentID string, exclusiveStartKey map[string]map[string]string) ([]byte, error) {
	payload := struct {
		TableName                 string                       `json:"TableName"`
		IndexName                 string                       `json:"IndexName"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
		ScanIndexForward          bool                         `json:"ScanIndexForward"`
		Limit                     int                          `json:"Limit"`
	}{
		TableName:              s.tableName,
		IndexName:              dynamoDBAssignmentQueueIndex,
		KeyConditionExpression: "agent_id = :agent_id",
		ExpressionAttributeValues: map[string]map[string]string{
			":agent_id": {"S": agentID},
		},
		ExclusiveStartKey: exclusiveStartKey,
		ScanIndexForward:  true,
		Limit:             dynamoDBAssignmentQueryLimit,
	}
	return json.Marshal(payload)
}

func assignmentOperationRecordFromDynamoDBItem(item map[string]map[string]string) (AssignmentOperationRecord, error) {
	var assignment Assignment
	if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "assignment_json")), &assignment); err != nil {
		return AssignmentOperationRecord{}, fmt.Errorf("decode assignment operation assignment_json: %w", err)
	}
	var events []TaskEvent
	if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "events_json")), &events); err != nil {
		return AssignmentOperationRecord{}, fmt.Errorf("decode assignment operation events_json: %w", err)
	}
	recordedAt, err := time.Parse(time.RFC3339Nano, dynamoDBStringValue(item, "recorded_at"))
	if err != nil {
		return AssignmentOperationRecord{}, fmt.Errorf("decode assignment operation recorded_at: %w", err)
	}
	record := AssignmentOperationRecord{
		SchemaVersion: dynamoDBStringValue(item, "schema_version"),
		OperationID:   dynamoDBStringValue(item, "operation_id"),
		OperationType: AssignmentOperationType(dynamoDBStringValue(item, "operation_type")),
		TaskID:        dynamoDBStringValue(item, "task_id"),
		AssignmentID:  dynamoDBStringValue(item, "assignment_id"),
		AgentID:       dynamoDBStringValue(item, "operation_agent_id"),
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    recordedAt,
	}
	if record.AgentID == "" {
		record.AgentID = dynamoDBStringValue(item, "agent_id")
	}
	if err := validateAssignmentOperationRecord(record); err != nil {
		return AssignmentOperationRecord{}, err
	}
	return record, nil
}

func assignmentProjectionFromDynamoDBItem(item map[string]map[string]string) (assignmentProjectionRecord, error) {
	if schema := dynamoDBStringValue(item, "schema_version"); schema != AssignmentProjectionSchemaVersion {
		return assignmentProjectionRecord{}, fmt.Errorf("unsupported assignment projection schema_version %q", schema)
	}
	var assignment Assignment
	if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "assignment_json")), &assignment); err != nil {
		return assignmentProjectionRecord{}, fmt.Errorf("decode assignment projection assignment_json: %w", err)
	}
	if assignment.ID == "" {
		return assignmentProjectionRecord{}, errors.New("assignment projection assignment_id is required")
	}
	lastEventSeq, err := strconv.ParseInt(dynamoDBNumberValue(item, "last_event_seq"), 10, 64)
	if err != nil {
		return assignmentProjectionRecord{}, fmt.Errorf("decode assignment projection last_event_seq: %w", err)
	}
	return assignmentProjectionRecord{Assignment: assignment, LastEventSeq: lastEventSeq}, nil
}

func agentActiveAssignmentFromDynamoDBItem(item map[string]map[string]string) (AssignmentActiveLease, error) {
	if schema := dynamoDBStringValue(item, "schema_version"); schema != AssignmentAgentActiveSchemaVersion {
		return AssignmentActiveLease{}, fmt.Errorf("unsupported agent active assignment schema_version %q", schema)
	}
	var heartbeatAt time.Time
	if raw := dynamoDBStringValue(item, "lease_heartbeat_at"); raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return AssignmentActiveLease{}, fmt.Errorf("decode agent active assignment lease_heartbeat_at: %w", err)
		}
		heartbeatAt = parsed
	}
	var expiresAt time.Time
	if raw := dynamoDBStringValue(item, "lease_expires_at"); raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return AssignmentActiveLease{}, fmt.Errorf("decode agent active assignment lease_expires_at: %w", err)
		}
		expiresAt = parsed
	}
	var expiresUnixMS int64
	if raw := dynamoDBNumberValue(item, "lease_expires_unix_ms"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return AssignmentActiveLease{}, fmt.Errorf("decode agent active assignment lease_expires_unix_ms: %w", err)
		}
		expiresUnixMS = parsed
	}
	lease := AssignmentActiveLease{
		AgentID:            dynamoDBStringValue(item, "agent_id"),
		ActiveAssignmentID: dynamoDBStringValue(item, "active_assignment_id"),
		LeaseToken:         dynamoDBStringValue(item, "lease_token"),
		HeartbeatAt:        heartbeatAt,
		LeaseExpiresAt:     expiresAt,
		LeaseExpiresUnixMS: expiresUnixMS,
	}
	if lease.AgentID == "" || lease.ActiveAssignmentID == "" {
		return AssignmentActiveLease{}, errors.New("agent active assignment agent_id and active_assignment_id are required")
	}
	return lease, nil
}

func dynamoDBStringValue(item map[string]map[string]string, key string) string {
	if item == nil {
		return ""
	}
	return item[key]["S"]
}

func dynamoDBNumberValue(item map[string]map[string]string, key string) string {
	if item == nil {
		return ""
	}
	return item[key]["N"]
}

func dynamoDBAssignmentProjectionPK(assignmentID string) string {
	return "ASSIGNMENT#" + assignmentID
}

func dynamoDBAgentActivePK(agentID string) string {
	return "AGENT#" + agentID
}

func assignmentOperationSortKey(record AssignmentOperationRecord) string {
	return record.RecordedAt.UTC().Format("20060102T150405.000000000Z") + "#" + fmt.Sprintf("%020d", assignmentOperationLastEventSeq(record)) + "#" + record.OperationID
}

var _ AssignmentOperationStore = (*DynamoDBAssignmentOperationStore)(nil)
var _ AssignmentOperationLoader = (*DynamoDBAssignmentOperationStore)(nil)
var _ AssignmentQueueReader = (*DynamoDBAssignmentOperationStore)(nil)
var _ AssignmentClaimer = (*DynamoDBAssignmentOperationStore)(nil)
var _ AssignmentActiveLeaseStore = (*DynamoDBAssignmentOperationStore)(nil)
var _ AssignmentProjectionReader = (*DynamoDBAssignmentOperationStore)(nil)
