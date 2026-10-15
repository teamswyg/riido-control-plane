package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

// Reads use their caller's deadline outside the mutation actor, and never query the global journal or queue GSI.
func (s *DynamoDBAssignmentOperationStore) LoadAgentAssignmentOperations(ctx context.Context, agentID string) ([]AssignmentOperationRecord, error) {
	if !s.enableAgentHistory {
		return nil, nil
	}
	credentials, err := s.credentialsProvider.Credentials(ctx)
	if err != nil {
		return nil, err
	}
	var records []AssignmentOperationRecord
	var start map[string]map[string]string
	for {
		payload, err := json.Marshal(struct {
			TableName                 string
			ConsistentRead            bool
			KeyConditionExpression    string
			ExpressionAttributeValues map[string]map[string]string
			ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
		}{s.tableName, true, "pk = :pk", map[string]map[string]string{":pk": {"S": agentHistoryPK(agentID)}}, start})
		if err != nil {
			return nil, err
		}
		body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBQueryTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
		if err != nil {
			return nil, err
		}
		var response struct {
			Items            []map[string]map[string]string
			LastEvaluatedKey map[string]map[string]string
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}
		for _, item := range response.Items {
			op, err := assignmentOperationRecordFromDynamoDBItem(item)
			if err != nil {
				return nil, err
			}
			if op.AgentID != agentID || op.Assignment.AgentID != agentID {
				return nil, fmt.Errorf("agent history scope mismatch")
			}
			records = append(records, op)
		}
		if len(response.LastEvaluatedKey) == 0 {
			return records, nil
		}
		start = response.LastEvaluatedKey
	}
}
