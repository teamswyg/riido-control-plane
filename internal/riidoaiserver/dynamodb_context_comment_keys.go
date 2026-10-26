package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) contextCommentKeys(ctx context.Context, id string) ([]map[string]map[string]string, error) {
	var items []map[string]map[string]string
	var start map[string]map[string]string
	seenItems, seenCursors := map[string]bool{}, map[string]bool{}
	for range assignmentContextMaxCommentPages + 1 {
		body, err := s.assignmentContextRequest(ctx, dynamoDBQueryTarget, struct {
			TableName                 string
			ConsistentRead            bool
			KeyConditionExpression    string
			ExpressionAttributeValues map[string]map[string]string
			ProjectionExpression      string
			Limit                     int
			ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
		}{s.tableName, true, "pk = :pk AND begins_with(sk, :read)", map[string]map[string]string{":pk": {"S": dynamoDBAssignmentContextPK(id)}, ":read": {"S": dynamoDBAssignmentContextCommentPrefix}}, "pk, sk", assignmentContextMaxCommentPages + 1, start})
		if err != nil {
			return nil, err
		}
		var response struct {
			Items            []map[string]map[string]string
			LastEvaluatedKey map[string]map[string]string
		}
		if err = json.Unmarshal(body, &response); err != nil {
			return nil, err
		}
		for _, item := range response.Items {
			key := item["sk"]["S"]
			if !validContextCommentKey(item, id) || seenItems[key] {
				return nil, errors.New("context comment scope or duplicate mismatch")
			}
			seenItems[key] = true
			items = append(items, item)
		}
		if len(items) > assignmentContextMaxCommentPages {
			return nil, errors.New("context comment page limit exceeded")
		}
		if len(response.LastEvaluatedKey) == 0 {
			return items, nil
		}
		previous := start["sk"]["S"]
		start = response.LastEvaluatedKey
		key := start["sk"]["S"]
		if !validContextCommentKey(start, id) || seenCursors[key] || key <= previous {
			return nil, errors.New("context comment cursor mismatch")
		}
		seenCursors[key] = true
	}
	return nil, errors.New("context comment query retry limit exceeded")
}

func validContextCommentKey(item map[string]map[string]string, id string) bool {
	if item["pk"]["S"] != dynamoDBAssignmentContextPK(id) {
		return false
	}
	for page := range assignmentContextMaxCommentPages {
		if item["sk"]["S"] == dynamoDBAssignmentContextCommentSK(page) {
			return true
		}
	}
	return false
}
