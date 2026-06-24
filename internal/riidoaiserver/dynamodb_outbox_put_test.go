package riidoaiserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDynamoDBOutboxWritesPutItem(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		requests <- capturedDynamoDBRequest{method: r.Method, header: r.Header.Clone(), body: body}
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	outbox, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-server-event-outbox",
		Endpoint:            server.URL,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKIDEXAMPLE", "SESSION"),
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBOutbox: %v", err)
	}
	defer outbox.Close()

	event := TaskEvent{
		Seq:          7,
		TaskID:       "task-a",
		AssignmentID: "assignment-1",
		AgentID:      "jykim1",
		Type:         EventAssignmentLeased,
		State:        AssignmentLeased,
		Message:      "leased",
		Metadata:     map[string]string{"lease_token": "lease-1"},
		At:           fixedNow,
	}
	if err := outbox.AppendTaskEvent(context.Background(), event); err != nil {
		t.Fatalf("AppendTaskEvent: %v", err)
	}

	got := <-requests
	assertDynamoDBOutboxPutRequest(t, got)
	assertDynamoDBOutboxPutPayload(t, decodeDynamoDBOutboxPutPayload(t, got.body))
}
