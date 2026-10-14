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

func TestDynamoDBAssignmentOperationStoreQueriesAgentQueue(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	queryRequests := make(chan []byte, 1)
	assignment := sampleQueuedAssignmentOperationRecord(fixedNow).Assignment
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if r.Header.Get("X-Amz-Target") != dynamoDBQueryTarget {
			t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
		}
		queryRequests <- body
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		if err := json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{item}}); err != nil {
			t.Errorf("encode response: %v", err)
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

	assignments, err := store.LoadAgentQueueAssignments(context.Background(), "jykim1")
	if err != nil {
		t.Fatalf("LoadAgentQueueAssignments: %v", err)
	}
	if len(assignments) != 1 || assignments[0].ID != assignment.ID || assignments[0].State != AssignmentQueued {
		t.Fatalf("assignments = %+v", assignments)
	}
	var payload struct {
		TableName                 string                       `json:"TableName"`
		IndexName                 string                       `json:"IndexName"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		ScanIndexForward          bool                         `json:"ScanIndexForward"`
		Limit                     int                          `json:"Limit"`
	}
	if err := json.Unmarshal(<-queryRequests, &payload); err != nil {
		t.Fatalf("decode query payload: %v", err)
	}
	if payload.TableName != "assignments" || payload.IndexName != dynamoDBAssignmentQueueIndex || !payload.ScanIndexForward || payload.Limit != dynamoDBAssignmentQueryLimit {
		t.Fatalf("query payload = %+v", payload)
	}
	if payload.KeyConditionExpression != "agent_id = :agent_id" {
		t.Fatalf("key condition = %q", payload.KeyConditionExpression)
	}
	assertDynamoDBString(t, payload.ExpressionAttributeValues, ":agent_id", "jykim1")
}
