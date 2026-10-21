package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *DynamoDBAssignmentOperationStore) loadAgentActiveAssignment(ctx context.Context, agentID string, credentials AWSCredentials) (AssignmentActiveLease, bool, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{
		TableName:      s.tableName,
		ConsistentRead: true,
		Key: map[string]map[string]string{
			"pk": {"S": dynamoDBAgentActivePK(agentID)},
			"sk": {"S": dynamoDBAgentActiveSK},
		},
	})
	if err != nil {
		return AssignmentActiveLease{}, false, err
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
		return AssignmentActiveLease{}, false, fmt.Errorf("dynamodb get agent active assignment: %w", err)
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return AssignmentActiveLease{}, false, fmt.Errorf("decode DynamoDB agent active assignment response: %w", err)
	}
	if len(response.Item) == 0 {
		return AssignmentActiveLease{}, false, nil
	}
	lease, err := agentActiveAssignmentFromDynamoDBItem(response.Item)
	if err != nil {
		return AssignmentActiveLease{}, false, err
	}
	return lease, true, nil
}
