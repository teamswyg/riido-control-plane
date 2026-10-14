package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreClaimContentionReturnsNoClaim(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 2)
	assignment := sampleQueuedAssignmentOperationRecord(fixedNow).Assignment
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 1)
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
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBQueryTarget:
			if err := json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{item}}); err != nil {
				t.Errorf("encode query response: %v", err)
			}
		case dynamoDBTransactWriteTarget:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#TransactionCanceledException","message":"condition failed"}`))
		default:
			t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
			w.WriteHeader(http.StatusBadRequest)
		}
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

	claim, err := store.ClaimNextAssignment(context.Background(), "jykim1", fixedNow)
	if err != nil {
		t.Fatalf("ClaimNextAssignment contention: %v", err)
	}
	if claim.Claimed {
		t.Fatalf("claim should be empty after contention: %+v", claim)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBQueryTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBTransactWriteTarget)
}
