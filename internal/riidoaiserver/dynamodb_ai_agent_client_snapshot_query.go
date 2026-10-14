package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *DynamoDBAIAgentClientSnapshot) doSnapshotQuery(ctx context.Context, credentials AWSCredentials) ([]byte, error) {
	payload, err := json.Marshal(struct {
		TableName                 string                       `json:"TableName"`
		ConsistentRead            bool                         `json:"ConsistentRead"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	}{
		TableName:              s.tableName,
		ConsistentRead:         false,
		KeyConditionExpression: "pk = :pk",
		ExpressionAttributeValues: map[string]map[string]string{
			":pk": {"S": dynamoDBAIAgentClientSnapshotPK},
		},
	})
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
		traceAttrs:   s.snapshotLoadTraceAttrs(),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb load AI Agent client snapshot: %w", err)
	}
	return body, nil
}
