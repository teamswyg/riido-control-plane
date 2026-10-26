package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) ListAssignmentContextReads(ctx context.Context, assignmentID string) ([]AssignmentContextRead, error) {
	var reads []AssignmentContextRead
	var start map[string]map[string]string
	for {
		body, err := s.assignmentContextRequest(ctx, dynamoDBQueryTarget, struct {
			TableName                 string
			ConsistentRead            bool
			KeyConditionExpression    string
			ExpressionAttributeValues map[string]map[string]string
			ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
		}{s.tableName, true, "pk = :pk AND begins_with(sk, :read)", map[string]map[string]string{
			":pk": {"S": dynamoDBAssignmentContextPK(assignmentID)}, ":read": {"S": dynamoDBAssignmentContextReadPrefix},
		}, start})
		if err != nil {
			return nil, err
		}
		var response struct {
			Items            []map[string]map[string]string
			LastEvaluatedKey map[string]map[string]string
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}
		for _, item := range response.Items {
			var read AssignmentContextRead
			if err := json.Unmarshal([]byte(item["payload"]["S"]), &read); err != nil || read.AssignmentID != assignmentID {
				return nil, errors.New("assignment context read scope mismatch")
			}
			reads = append(reads, read)
		}
		if len(response.LastEvaluatedKey) == 0 {
			return reads, nil
		}
		start = response.LastEvaluatedKey
	}
}
