package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) contextRetentionItem(ctx context.Context, id, sk string, fields ...string) (map[string]map[string]string, error) {
	key := map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK(id)}, "sk": {"S": sk}}
	projection := ""
	if len(fields) > 0 {
		projection = fields[0]
	}
	body, err := s.assignmentContextRequest(ctx, dynamoDBGetItemTarget, struct {
		TableName            string
		ConsistentRead       bool
		Key                  map[string]map[string]string
		ProjectionExpression string `json:"ProjectionExpression,omitempty"`
	}{s.tableName, true, key, projection})
	if err != nil {
		return nil, err
	}
	var response struct{ Item map[string]map[string]string }
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if len(response.Item) > 0 && (response.Item["pk"]["S"] != key["pk"]["S"] || response.Item["sk"]["S"] != sk) {
		return nil, errors.New("context lifecycle scope mismatch")
	}
	return response.Item, nil
}

func (s *DynamoDBAssignmentOperationStore) contextRetentionPayloads(ctx context.Context, id string) ([]map[string]map[string]string, error) {
	snapshot, err := s.contextRetentionItem(ctx, id, dynamoDBAssignmentContextSnapshotSK, "pk, sk")
	if err != nil {
		return nil, err
	}
	items := []map[string]map[string]string{}
	if len(snapshot) > 0 {
		items = append(items, snapshot)
	}
	comments, err := s.contextCommentKeys(ctx, id)
	if err != nil {
		return nil, err
	}
	return append(items, comments...), nil
}
