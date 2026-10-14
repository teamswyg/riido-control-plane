package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (s *DynamoDBAIAgentClientSnapshot) readTaskAgentGeneration(ctx context.Context, scope taskAgentGenerationScope, credentials AWSCredentials) (string, error) {
	payload, err := json.Marshal(struct {
		TableName      string
		ConsistentRead bool
		Key            map[string]map[string]string
	}{s.tableName, true, taskAgentDynamoKey(scope.taskID, scope.agentID, taskAgentGenerationSK)})
	if err != nil {
		return "", err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBGetItemTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err != nil {
		return "", err
	}
	var response struct{ Item map[string]map[string]string }
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	return dynamoDBStringValue(response.Item, "generation"), nil
}
