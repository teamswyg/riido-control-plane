package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (s *DynamoDBAssignmentOperationStore) executionTerminationGet(ctx context.Context, pk, sk string) (map[string]map[string]string, error) {
	body, err := s.assignmentContextRequest(ctx, dynamoDBGetItemTarget, struct {
		TableName      string
		ConsistentRead bool
		Key            map[string]map[string]string
	}{s.tableName, true, map[string]map[string]string{"pk": {"S": pk}, "sk": {"S": sk}}})
	if err != nil {
		return nil, err
	}
	var response struct{ Item map[string]map[string]string }
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response.Item, nil
}
