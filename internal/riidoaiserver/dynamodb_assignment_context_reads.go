package riidoaiserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// AppendAssignmentContextReads writes each batch atomically. The sort key
// orders records by time; the random suffix keeps concurrent reads distinct.
func (s *DynamoDBAssignmentOperationStore) AppendAssignmentContextReads(ctx context.Context, reads []AssignmentContextRead) error {
	for start := 0; start < len(reads); start += dynamoDBAssignmentContextReadBatch {
		batch := reads[start:min(start+dynamoDBAssignmentContextReadBatch, len(reads))]
		writes := make([]dynamoDBTransactWritePut, 0, len(batch))
		for i, read := range batch {
			if read.AssignmentID == "" {
				return errors.New("assignment_id is required")
			}
			encoded, err := json.Marshal(read)
			if err != nil {
				return err
			}
			var suffix [6]byte
			if _, err := rand.Read(suffix[:]); err != nil {
				return err
			}
			var write dynamoDBTransactWritePut
			write.Put.TableName = s.tableName
			write.Put.ConditionExpression = "attribute_not_exists(pk) AND attribute_not_exists(sk)"
			write.Put.Item = map[string]map[string]string{
				"pk":      {"S": dynamoDBAssignmentContextPK(read.AssignmentID)},
				"sk":      {"S": fmt.Sprintf("%s%s#%04d#%s", dynamoDBAssignmentContextReadPrefix, read.ServedAt.UTC().Format(dynamoDBAssignmentContextTimeFormat), start+i, hex.EncodeToString(suffix[:]))},
				"schema":  {"S": "assignment_context_read/v1"},
				"payload": {"S": string(encoded)},
			}
			writes = append(writes, write)
		}
		if _, err := s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct {
			TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
		}{writes}); err != nil {
			return err
		}
	}
	return nil
}
