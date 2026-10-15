package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

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
