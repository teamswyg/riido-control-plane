package riidoaiserver

import "encoding/json"

func (s *DynamoDBAssignmentOperationStore) executionTerminationActions(id executionTerminationIdentity, set executionTerminationReadSet) ([]dynamoDBTransactWriteItem, error) {
	claimCheck := &dynamoDBTransactWriteConditionCheckAction{
		TableName: s.tableName, Key: executionTerminationItemKey(set.claim),
		ConditionExpression:       "#schema = :schema AND payload = :payload",
		ExpressionAttributeNames:  map[string]string{"#schema": "schema"},
		ExpressionAttributeValues: map[string]map[string]string{":schema": set.claim["schema"], ":payload": set.claim["payload"]},
	}
	projectionCheck := &dynamoDBTransactWriteConditionCheckAction{
		TableName: s.tableName, Key: executionTerminationItemKey(set.projection),
		ConditionExpression:       "schema_version = :schema AND assignment_state = :state AND last_event_seq = :seq AND lease_epoch = :epoch",
		ExpressionAttributeValues: map[string]map[string]string{":schema": set.projection["schema_version"], ":state": set.projection["assignment_state"], ":seq": set.projection["last_event_seq"], ":epoch": set.projection["lease_epoch"]},
	}
	actions := []dynamoDBTransactWriteItem{{ConditionCheck: claimCheck}, {ConditionCheck: projectionCheck}}
	if len(set.receipt) > 0 {
		return append(actions, dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: executionTerminationItemKey(set.receipt), ConditionExpression: "#schema = :schema AND payload = :payload", ExpressionAttributeNames: map[string]string{"#schema": "schema"}, ExpressionAttributeValues: map[string]map[string]string{":schema": set.receipt["schema"], ":payload": set.receipt["payload"]}}}), nil
	}
	payload, err := json.Marshal(set.value)
	if err != nil {
		return nil, err
	}
	item := map[string]map[string]string{"pk": {"S": "DEVICE_EXECUTION_TERMINATION#" + id.deviceID}, "sk": {"S": id.key()}, "schema": {"S": executionTerminationItemSchema}, "payload": {"S": string(payload)}}
	return append(actions, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)", Item: item}}), nil
}

func executionTerminationItemKey(item map[string]map[string]string) map[string]map[string]string {
	return map[string]map[string]string{"pk": item["pk"], "sk": item["sk"]}
}
