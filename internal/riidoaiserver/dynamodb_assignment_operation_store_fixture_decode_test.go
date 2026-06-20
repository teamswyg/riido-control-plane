package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func decodeDynamoDBPutPayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBPutPayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBPutItemTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBPutPayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	return payload
}

func decodeDynamoDBDeletePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBDeletePayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBDeleteItemTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBDeletePayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode delete payload: %v", err)
	}
	return payload
}

func decodeDynamoDBUpdatePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBUpdatePayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBUpdateItemTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBUpdatePayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode update payload: %v", err)
	}
	return payload
}

func decodeDynamoDBTransactWritePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBTransactWritePayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBTransactWriteTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBTransactWritePayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode transact payload: %v", err)
	}
	return payload
}

func decodeDynamoDBRepairTransactWritePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBRepairTransactWritePayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBTransactWriteTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBRepairTransactWritePayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode repair transact payload: %v", err)
	}
	return payload
}
