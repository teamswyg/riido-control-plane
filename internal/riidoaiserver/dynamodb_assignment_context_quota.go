package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

const assignmentContextReadQuotaSK = "READ_QUOTA"

// A single conditional UpdateItem admits at most 100 attempts across servers.
// Request IDs intentionally are not deduplicated: repeating a live read is an
// additional attempt. Terminal transactions fence this row with payload TTL.
func (s *DynamoDBAssignmentOperationStore) ConsumeAssignmentContextRead(ctx context.Context, assignmentID string) (bool, error) {
	id := strings.TrimSpace(assignmentID)
	if id == "" {
		return false, errors.New("assignment_id is required")
	}
	key, err := assignmentContextQuotaKey(id)
	if err != nil {
		return false, err
	}
	_, err = s.assignmentContextRequest(ctx, dynamoDBUpdateItemTarget, struct {
		TableName                             string
		Key                                   map[string]map[string]string
		UpdateExpression, ConditionExpression string
		ExpressionAttributeNames              map[string]string
		ExpressionAttributeValues             map[string]map[string]string
	}{
		TableName:                 s.tableName,
		Key:                       key,
		UpdateExpression:          "SET #count = if_not_exists(#count, :zero) + :one",
		ConditionExpression:       "attribute_not_exists(#ttl) AND (attribute_not_exists(#count) OR #count < :limit)",
		ExpressionAttributeNames:  map[string]string{"#count": "read_attempts", "#ttl": assignmentContextTTLAttribute},
		ExpressionAttributeValues: map[string]map[string]string{":zero": {"N": "0"}, ":one": {"N": "1"}, ":limit": {"N": strconv.Itoa(assignmentContextReadLimit)}},
	})
	if err == nil {
		return true, nil
	}
	var apiErr dynamoDBAPIError
	if errors.As(err, &apiErr) {
		var detail struct {
			Type string `json:"__type"`
		}
		if json.Unmarshal(apiErr.body, &detail) == nil && strings.HasSuffix(detail.Type, "ConditionalCheckFailedException") {
			return false, nil
		}
	}
	return false, err
}
