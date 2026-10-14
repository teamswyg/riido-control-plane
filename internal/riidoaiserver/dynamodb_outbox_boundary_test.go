package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDynamoDBOutboxNilAndClosedBoundaries(t *testing.T) {
	var nilOutbox *DynamoDBOutbox
	var nilContext context.Context
	if err := nilOutbox.AppendTaskEvent(nilContext, TaskEvent{}); err != nil {
		t.Fatalf("nil AppendTaskEvent() error = %v, want nil", err)
	}
	if err := nilOutbox.Close(); err != nil {
		t.Fatalf("nil Close() error = %v, want nil", err)
	}

	outbox := newDynamoDBOutboxForBoundary(t, http.StatusOK, `{}`)
	if err := outbox.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := outbox.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	err := outbox.AppendTaskEvent(context.Background(), TaskEvent{})
	if err == nil || err.Error() != "riidoaiserver: DynamoDB outbox closed" {
		t.Fatalf("AppendTaskEvent closed error = %v", err)
	}
}

func TestDynamoDBOutboxLoopRejectsNilEventCommand(t *testing.T) {
	outbox := newDynamoDBOutboxForBoundary(t, http.StatusOK, `{}`)
	defer outbox.Close()
	reply := make(chan error, 1)
	outbox.commands <- dynamoDBOutboxCommand{
		ctx:   context.Background(),
		reply: reply,
	}
	if err := <-reply; err == nil || err.Error() != "riidoaiserver: nil DynamoDB outbox event" {
		t.Fatalf("nil outbox command error = %v", err)
	}
}

func newDynamoDBOutboxForBoundary(t *testing.T, status int, body string) *DynamoDBOutbox {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	outbox, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		Region:              "ap-northeast-2",
		TableName:           "events",
		Endpoint:            server.URL,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKID", ""),
	})
	if err != nil {
		t.Fatalf("NewDynamoDBOutbox: %v", err)
	}
	return outbox
}
