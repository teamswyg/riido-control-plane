package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *DynamoDBAIAgentClientSnapshot) doSnapshotQuery(ctx context.Context, credentials AWSCredentials) ([]byte, error) {
	var cursor map[string]map[string]string
	var items []map[string]map[string]string
	for {
		payload, err := json.Marshal(struct {
			ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
			TableName                 string                       `json:"TableName"`
			ConsistentRead            bool                         `json:"ConsistentRead"`
			KeyConditionExpression    string                       `json:"KeyConditionExpression"`
			ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		}{
			ExclusiveStartKey:      cursor,
			TableName:              s.tableName,
			ConsistentRead:         true,
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
		var page struct {
			Items            []map[string]map[string]string
			LastEvaluatedKey map[string]map[string]string
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		items = append(items, page.Items...)
		if len(page.LastEvaluatedKey) == 0 {
			return json.Marshal(struct {
				Items []map[string]map[string]string
			}{items})
		}
		cursor = page.LastEvaluatedKey
	}
}
