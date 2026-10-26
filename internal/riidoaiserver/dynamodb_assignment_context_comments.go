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
	if strings.TrimSpace(assignmentID) == "" || len(pages) > assignmentContextMaxCommentPages {
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
	if s.enableAssignmentContextRetention {
		items := make([]map[string]map[string]string, 0, len(writes))
		for _, write := range writes {
			items = append(items, write.Put.Item)
		}
		return s.saveContextPayloads(ctx, assignmentID, items)
	}
	_, err := s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct {
		TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
	}{writes})
	return err
}
