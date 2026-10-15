package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

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
