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

func TestDynamoDBAssignmentOperationStoreLoadsAssignmentProjection(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 1)
	assignment := sampleAssignmentOperationRecord(fixedNow).Assignment
	assignment.State = AssignmentCancelling
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 7)
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
		_ = json.NewEncoder(w).Encode(map[string]any{"Item": item})
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

	projection, ok, err := store.LoadAssignmentProjection(context.Background(), assignment.ID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection: %v", err)
	}
	if !ok || projection.Assignment.ID != assignment.ID || projection.Assignment.State != AssignmentCancelling || projection.LastEventSeq != 7 {
		t.Fatalf("projection ok=%v value=%+v", ok, projection)
	}
	got := <-requests
	assertDynamoDBTarget(t, got, dynamoDBGetItemTarget)
	var payload struct {
		TableName string                       `json:"TableName"`
		Key       map[string]map[string]string `json:"Key"`
	}
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode get payload: %v", err)
	}
	assertDynamoDBString(t, payload.Key, "pk", "ASSIGNMENT#asn-000001")
	assertDynamoDBString(t, payload.Key, "sk", dynamoDBAssignmentProjectionSK)
}
