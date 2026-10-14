package riidoaiserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreReleasesAgentActiveAssignmentOnTerminalEvent(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		requests <- capturedDynamoDBRequest{
			method: r.Method,
			header: r.Header.Clone(),
			body:   body,
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAssignmentOperationStore: %v", err)
	}
	defer store.Close()

	record := sampleTerminalAssignmentOperationRecord(fixedNow)
	if err := store.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatalf("SaveAssignmentOperation terminal: %v", err)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	deletePayload := decodeDynamoDBDeletePayload(t, <-requests)
	if deletePayload.TableName != "assignments" {
		t.Fatalf("delete table = %q", deletePayload.TableName)
	}
	if deletePayload.ConditionExpression != "active_assignment_id = :assignment_id" {
		t.Fatalf("delete condition = %q", deletePayload.ConditionExpression)
	}
	assertDynamoDBString(t, deletePayload.Key, "pk", "AGENT#jykim1")
	assertDynamoDBString(t, deletePayload.Key, "sk", dynamoDBAgentActiveSK)
	assertDynamoDBString(t, deletePayload.ExpressionAttributeValues, ":assignment_id", "asn-000001")
}
