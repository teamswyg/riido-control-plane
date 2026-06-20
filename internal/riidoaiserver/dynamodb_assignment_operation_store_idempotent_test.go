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

func TestDynamoDBAssignmentOperationStoreTreatsConditionalCheckFailedAsIdempotent(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		var payload struct {
			Item map[string]map[string]string `json:"Item"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode request: %v", err)
		}
		pk := payload.Item["pk"]["S"]
		requests <- pk
		if pk == dynamoDBAssignmentOperationPK {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException","message":"duplicate"}`))
			return
		}
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
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAssignmentOperationStore: %v", err)
	}
	defer store.Close()

	err = store.SaveAssignmentOperation(context.Background(), sampleAssignmentOperationRecord(time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)))
	if err != nil {
		t.Fatalf("SaveAssignmentOperation duplicate: %v", err)
	}
	if first, second := <-requests, <-requests; first != dynamoDBAssignmentOperationPK || second != "ASSIGNMENT#asn-000001" {
		t.Fatalf("request pks = %q, %q", first, second)
	}
}
