package riidoaiserver

import (
	"context"
	"encoding/json"
)

func sameTaskAgentCreation(a, b Assignment) bool {
	return a.ID == b.ID && a.TaskID == b.TaskID && a.AgentID == b.AgentID && a.CreatedAt.Equal(b.CreatedAt) && sameRecoveryPolicy(a.RecoveryPolicy, b.RecoveryPolicy)
}

func (s *DynamoDBAssignmentOperationStore) getTaskAgentItem(ctx context.Context, key map[string]map[string]string, credentials AWSCredentials) (map[string]map[string]string, bool, error) {
	payload, err := json.Marshal(struct {
		TableName      string
		ConsistentRead bool
		Key            map[string]map[string]string
	}{s.tableName, true, key})
	if err != nil {
		return nil, false, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBGetItemTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err != nil {
		return nil, false, err
	}
	var response struct{ Item map[string]map[string]string }
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, false, err
	}
	return response.Item, len(response.Item) > 0, nil
}
