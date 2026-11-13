package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *DynamoDBAssignmentOperationStore) loadAssignmentProjection(ctx context.Context, assignmentID string, credentials AWSCredentials) (assignmentProjectionRecord, bool, error) {
	return s.readAssignmentProjection(ctx, assignmentID, credentials, false)
}

func (s *DynamoDBAssignmentOperationStore) readAssignmentProjection(ctx context.Context, assignmentID string, credentials AWSCredentials, readTerminal bool) (assignmentProjectionRecord, bool, error) {
	item, err := s.getAssignmentJournalItem(ctx, credentials, dynamoDBAssignmentProjectionPK(assignmentID), dynamoDBAssignmentProjectionSK)
	if err != nil {
		return assignmentProjectionRecord{}, false, fmt.Errorf("dynamodb get assignment projection: %w", err)
	}
	if len(item) == 0 {
		return assignmentProjectionRecord{}, false, nil
	}
	projection, err := assignmentProjectionFromDynamoDBItem(item)
	if err != nil {
		return assignmentProjectionRecord{}, false, err
	}
	if readTerminal && isTerminal(projection.Assignment.State) && AssignmentOperationType(dynamoDBStringValue(item, "operation_type")) == AssignmentOperationAgentEvent {
		projection.TerminalEvent, err = s.readTerminalAssignmentEvent(ctx, credentials, projection, item)
		if err != nil {
			return assignmentProjectionRecord{}, false, err
		}
		if projection.TerminalEvent != nil {
			projection.WorkspacePreparationFailed = isWorkspacePreparationFailure(projection.Assignment, *projection.TerminalEvent)
		}
	}
	return projection, true, nil
}

func (s *DynamoDBAssignmentOperationStore) getAssignmentJournalItem(ctx context.Context, credentials AWSCredentials, pk, sk string) (map[string]map[string]string, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key: map[string]map[string]string{
			"pk": {"S": pk},
			"sk": {"S": sk},
		},
	})
	if err != nil {
		return nil, err
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
		return nil, err
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode DynamoDB assignment journal response: %w", err)
	}
	return response.Item, nil
}
