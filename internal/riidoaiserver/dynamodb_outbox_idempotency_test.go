package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDynamoDBOutboxTreatsConditionalCheckFailedAsIdempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException","message":"duplicate"}`))
	}))
	defer server.Close()

	outbox, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		Region:              "ap-northeast-2",
		TableName:           "events",
		Endpoint:            server.URL,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKID", ""),
	})
	if err != nil {
		t.Fatalf("NewDynamoDBOutbox: %v", err)
	}
	defer outbox.Close()

	err = outbox.AppendTaskEvent(context.Background(), TaskEvent{
		Seq:    1,
		TaskID: "task-a",
		Type:   EventAssignmentQueued,
		At:     time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("AppendTaskEvent duplicate: %v", err)
	}
}
