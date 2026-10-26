package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const dynamoDBAssignmentContextCommentPrefix = "COMMENTS#"

var (
	_ AssignmentContextStore = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentContextStore = (*MemoryAssignmentContextStore)(nil)
)

func dynamoDBAssignmentContextCommentSK(page int) string {
	return fmt.Sprintf("%s%04d", dynamoDBAssignmentContextCommentPrefix, page)
}

// SaveAssignmentContextComments writes all pages in one transaction (never a partial set).
func (s *DynamoDBAssignmentOperationStore) SaveAssignmentContextComments(ctx context.Context, assignmentID string, pages [][]AIAgentTaskContextComment) error {
	if strings.TrimSpace(assignmentID) == "" || len(pages) > dynamoDBAssignmentContextReadBatch {
		return errors.New("assignment_id is required and pages must fit one transaction")
	}
	writes := make([]dynamoDBTransactWritePut, 0, len(pages))
	for i, page := range pages {
		encoded, err := json.Marshal(page)
		if err != nil {
			return err
		}
		compressed, err := gzipBase64(encoded)
		if err != nil {
			return err
		}
		var write dynamoDBTransactWritePut
		write.Put.TableName = s.tableName
		write.Put.Item = map[string]map[string]string{
			"pk":           {"S": dynamoDBAssignmentContextPK(assignmentID)},
			"sk":           {"S": dynamoDBAssignmentContextCommentSK(i)},
			"schema":       {"S": "assignment_context_comments/v1"},
			"payload_gzip": {"S": compressed},
		}
		writes = append(writes, write)
	}
	_, err := s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct {
		TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
	}{writes})
	return err
}

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
	decoded, err := gunzipBase64(response.Item["payload_gzip"]["S"])
	if err != nil {
		return nil, false, err
	}
	var comments []AIAgentTaskContextComment
	return comments, true, json.Unmarshal(decoded, &comments)
}
