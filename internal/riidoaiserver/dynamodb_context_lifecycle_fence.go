package riidoaiserver

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"strconv"
)

func (s *DynamoDBAssignmentOperationStore) contextLifecycleFence(id string, previous map[string]map[string]string, expiry int64) (dynamoDBTransactWriteItem, error) {
	key := map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK(id)}, "sk": {"S": assignmentContextLifecycleSK}}
	condition := "attribute_not_exists(pk)"
	var values map[string]map[string]string
	if len(previous) > 0 {
		version := previous["revision"]["S"]
		if version == "" {
			return dynamoDBTransactWriteItem{}, errors.New("invalid context lifecycle revision")
		}
		condition = "revision = :revision"
		values = map[string]map[string]string{":revision": {"S": version}}
	}
	key["revision"] = map[string]string{"S": rand.Text()}
	key["schema"] = map[string]string{"S": "assignment_context_lifecycle/v1"}
	if expiry > 0 {
		key[assignmentContextExpiryField] = map[string]string{"N": strconv.FormatInt(expiry, 10)}
	}
	encoded, err := json.Marshal(key)
	if err != nil || len(encoded) > contextLifecycleItemBytes {
		return dynamoDBTransactWriteItem{}, errors.New("context lifecycle byte budget exceeded")
	}
	return dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: key, ConditionExpression: condition, ExpressionAttributeValues: values}}, nil
}

func (s *DynamoDBAssignmentOperationStore) contextProjectionFence(id string, projection assignmentProjectionRecord, found bool) dynamoDBTransactWriteItem {
	check := &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: map[string]map[string]string{"pk": {"S": dynamoDBAssignmentProjectionPK(id)}, "sk": {"S": dynamoDBAssignmentProjectionSK}}, ConditionExpression: "attribute_not_exists(pk)"}
	if found {
		check.ConditionExpression = "assignment_state = :state AND last_event_seq = :seq"
		check.ExpressionAttributeValues = map[string]map[string]string{":state": {"S": string(projection.Assignment.State)}, ":seq": {"N": strconv.FormatInt(projection.LastEventSeq, 10)}}
	}
	return dynamoDBTransactWriteItem{ConditionCheck: check}
}
