package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (s *DynamoDBAssignmentOperationStore) loadTaskAgentDurableExecutions(ctx context.Context, taskID, agentID string) (string, []AssignmentProjection, error) {
	credentials, err := s.credentialsProvider.Credentials(ctx)
	if err != nil {
		return "", nil, err
	}
	marker, _, err := s.getTaskAgentItem(ctx, taskAgentDynamoKey(taskID, agentID, taskAgentGenerationSK), credentials)
	if err != nil {
		return "", nil, err
	}
	var assignments []AssignmentProjection
	var cursor map[string]map[string]string
	for {
		payload, err := json.Marshal(struct {
			TableName                 string
			ConsistentRead            bool
			KeyConditionExpression    string
			ExpressionAttributeValues map[string]map[string]string
			ExclusiveStartKey         map[string]map[string]string `json:",omitempty"`
		}{s.tableName, true, "pk = :pk AND begins_with(sk, :exec)", map[string]map[string]string{":pk": {"S": dynamoDBTaskAgentPK(taskID, agentID)}, ":exec": {"S": "EXEC#"}}, cursor})
		if err != nil {
			return "", nil, err
		}
		body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBQueryTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
		if err != nil {
			return "", nil, err
		}
		var response struct {
			Items            []map[string]map[string]string
			LastEvaluatedKey map[string]map[string]string
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return "", nil, err
		}
		for _, item := range response.Items {
			id := dynamoDBStringValue(item, "assignment_id")
			if id == "" {
				continue
			}
			projection, found, err := s.loadAssignmentProjection(ctx, id, credentials)
			if err != nil {
				return "", nil, err
			}
			if found && projection.Assignment.TaskID == taskID && projection.Assignment.AgentID == agentID && !isTerminal(projection.Assignment.State) {
				assignments = append(assignments, AssignmentProjection(projection))
			}
		}
		if len(response.LastEvaluatedKey) == 0 {
			break
		}
		cursor = response.LastEvaluatedKey
	}
	latest, _, err := s.getTaskAgentItem(ctx, taskAgentDynamoKey(taskID, agentID, taskAgentGenerationSK), credentials)
	if err != nil {
		return "", nil, err
	}
	if dynamoDBStringValue(latest, "generation") != dynamoDBStringValue(marker, "generation") {
		return "", nil, ErrAIAgentMembershipChanged
	}
	return dynamoDBStringValue(marker, "generation"), assignments, nil
}
