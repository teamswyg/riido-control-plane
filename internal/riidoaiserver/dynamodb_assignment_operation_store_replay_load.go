package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

func (s *DynamoDBAssignmentOperationStore) load(ctx context.Context, credentials AWSCredentials) ([]AssignmentOperationRecord, error) {
	var records []AssignmentOperationRecord
	var startKey map[string]map[string]string
	for {
		payload, err := s.queryPayload(startKey)
		if err != nil {
			return nil, err
		}
		body, err := loadDynamoDBAssignmentOperationPage(ctx, dynamoDBRequest{
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
			return nil, err
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
		if err := paceDynamoDBAssignmentOperationReplay(ctx); err != nil {
			return nil, err
		}
	}
}
