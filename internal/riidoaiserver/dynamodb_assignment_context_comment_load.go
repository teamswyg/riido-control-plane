package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (s *DynamoDBAssignmentOperationStore) LoadAssignmentContextCommentPage(ctx context.Context, assignmentID string, page int) ([]AIAgentTaskContextComment, bool, error) {
	if page < 0 {
		return nil, false, nil
	}
	body, err := s.assignmentContextRequest(ctx, dynamoDBGetItemTarget, struct {
		TableName      string
		ConsistentRead bool
		Key            map[string]map[string]string
	}{s.tableName, true, map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK(assignmentID)}, "sk": {"S": dynamoDBAssignmentContextCommentSK(page)}}})
	if err != nil {
		return nil, false, err
	}
	var response struct{ Item map[string]map[string]string }
	if err := json.Unmarshal(body, &response); err != nil || len(response.Item) == 0 {
		return nil, false, err
	}
	if expired, err := assignmentContextItemExpired(response.Item, s.now()); expired || err != nil {
		return nil, false, err
	}
	decoded, err := gunzipBase64(response.Item["payload_gzip"]["S"])
	if err != nil {
		return nil, false, err
	}
	var comments []AIAgentTaskContextComment
	return comments, true, json.Unmarshal(decoded, &comments)
}
