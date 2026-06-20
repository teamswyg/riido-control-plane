package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func decodeDynamoDBPutPayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBPutPayload {
	t.Helper()
	return decodeDynamoDBPayload[dynamoDBPutPayload](t, got, dynamoDBPutItemTarget, "payload")
}

func decodeDynamoDBDeletePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBDeletePayload {
	t.Helper()
	return decodeDynamoDBPayload[dynamoDBDeletePayload](t, got, dynamoDBDeleteItemTarget, "delete payload")
}

func decodeDynamoDBUpdatePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBUpdatePayload {
	t.Helper()
	return decodeDynamoDBPayload[dynamoDBUpdatePayload](t, got, dynamoDBUpdateItemTarget, "update payload")
}

func decodeDynamoDBTransactWritePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBTransactWritePayload {
	t.Helper()
	return decodeDynamoDBPayload[dynamoDBTransactWritePayload](t, got, dynamoDBTransactWriteTarget, "transact payload")
}

func decodeDynamoDBRepairTransactWritePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBRepairTransactWritePayload {
	t.Helper()
	return decodeDynamoDBPayload[dynamoDBRepairTransactWritePayload](t, got, dynamoDBTransactWriteTarget, "repair transact payload")
}

func decodeDynamoDBPayload[T any](
	t *testing.T,
	got capturedDynamoDBRequest,
	target string,
	label string,
) T {
	t.Helper()
	assertDynamoDBTarget(t, got, target)
	var payload T
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode %s: %v", label, err)
	}
	return payload
}
