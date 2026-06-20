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

func TestDynamoDBAssignmentOperationStoreQueriesReplayableJournal(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	queryRequests := make(chan []byte, 1)
	record := sampleAssignmentOperationRecord(fixedNow)
	item := sampleAssignmentOperationDynamoDBItem(t, record)
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

	records, err := store.LoadAssignmentOperations(context.Background())
	if err != nil {
		t.Fatalf("LoadAssignmentOperations: %v", err)
	}
	if len(records) != 1 || records[0].OperationID != record.OperationID || records[0].Assignment.ID != record.Assignment.ID {
		t.Fatalf("records = %+v", records)
	}
	var payload struct {
		TableName                 string                       `json:"TableName"`
		ConsistentRead            bool                         `json:"ConsistentRead"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		ScanIndexForward          bool                         `json:"ScanIndexForward"`
		Limit                     int                          `json:"Limit"`
	}
	if err := json.Unmarshal(<-queryRequests, &payload); err != nil {
		t.Fatalf("decode query payload: %v", err)
	}
	if payload.TableName != "assignments" || !payload.ConsistentRead || !payload.ScanIndexForward || payload.Limit != dynamoDBAssignmentQueryLimit {
		t.Fatalf("query payload = %+v", payload)
	}
	if payload.KeyConditionExpression != "pk = :pk" {
		t.Fatalf("key condition = %q", payload.KeyConditionExpression)
	}
	assertDynamoDBString(t, payload.ExpressionAttributeValues, ":pk", dynamoDBAssignmentOperationPK)
}
