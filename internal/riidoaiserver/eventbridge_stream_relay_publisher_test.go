package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEventBridgeStreamRelayPublisherPutsEvent(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 5, 6, 7, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 1)
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
		_, _ = w.Write([]byte(`{"FailedEntryCount":0,"Entries":[{"EventId":"eventbridge-1"}]}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "SESSION")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	publisher, err := NewEventBridgeStreamRelayPublisher(EventBridgeStreamRelayPublisherConfig{
		Region:              "ap-northeast-2",
		EventBusName:        "riido-ai-server-events",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewEventBridgeStreamRelayPublisher: %v", err)
	}
	defer publisher.Close()

	event := streamRelayEventForEventBridgeTest(fixedNow)
	if err := publisher.PublishStreamRelayEvent(context.Background(), event); err != nil {
		t.Fatalf("PublishStreamRelayEvent: %v", err)
	}

	got := <-requests
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("Content-Type") != eventBridgeJSONContentType {
		t.Fatalf("content-type = %q", got.header.Get("Content-Type"))
	}
	if got.header.Get("X-Amz-Target") != eventBridgePutEventsTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	if got.header.Get("X-Amz-Date") != "20260526T050607Z" {
		t.Fatalf("x-amz-date = %q", got.header.Get("X-Amz-Date"))
	}
	if got.header.Get("X-Amz-Security-Token") != "SESSION" {
		t.Fatalf("session token = %q", got.header.Get("X-Amz-Security-Token"))
	}
	authorization := got.header.Get("Authorization")
	if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20260526/ap-northeast-2/events/aws4_request") {
		t.Fatalf("authorization = %q", authorization)
	}
	if !strings.Contains(authorization, "SignedHeaders=content-type;host;x-amz-date;x-amz-security-token;x-amz-target") {
		t.Fatalf("authorization signed headers = %q", authorization)
	}

	var payload eventBridgePutEventsRequest
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if len(payload.Entries) != 1 {
		t.Fatalf("entries = %+v", payload.Entries)
	}
	entry := payload.Entries[0]
	if entry.Source != defaultEventBridgeSource || entry.DetailType != defaultEventBridgeDetailType || entry.EventBusName != "riido-ai-server-events" {
		t.Fatalf("entry = %+v", entry)
	}
	if len(entry.Resources) != 1 || entry.Resources[0] != event.StreamARN {
		t.Fatalf("resources = %+v", entry.Resources)
	}
	var detail StreamRelayEvent
	if err := json.Unmarshal([]byte(entry.Detail), &detail); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if detail.SchemaVersion != StreamRelayEventSchemaVersion || detail.Record.Event.TaskID != "task-a" || detail.SequenceNumber != "42" {
		t.Fatalf("detail = %+v", detail)
	}
}

func TestEventBridgeStreamRelayPublisherReportsEntryFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"FailedEntryCount":1,"Entries":[{"ErrorCode":"InternalFailure","ErrorMessage":"try again"}]}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	publisher, err := NewEventBridgeStreamRelayPublisher(EventBridgeStreamRelayPublisherConfig{
		Region:              "ap-northeast-2",
		EventBusName:        "riido-ai-server-events",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatalf("NewEventBridgeStreamRelayPublisher: %v", err)
	}
	defer publisher.Close()

	err = publisher.PublishStreamRelayEvent(context.Background(), streamRelayEventForEventBridgeTest(time.Date(2026, 5, 26, 5, 6, 7, 0, time.UTC)))
	if err == nil || !strings.Contains(err.Error(), "InternalFailure") {
		t.Fatalf("expected EventBridge failure, got %v", err)
	}
}

func TestEventBridgeStreamRelayPublisherRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	if _, err := NewEventBridgeStreamRelayPublisher(EventBridgeStreamRelayPublisherConfig{EventBusName: "bus", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing region error")
	}
	if _, err := NewEventBridgeStreamRelayPublisher(EventBridgeStreamRelayPublisherConfig{Region: "ap-northeast-2", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing bus error")
	}
	if _, err := NewEventBridgeStreamRelayPublisher(EventBridgeStreamRelayPublisherConfig{Region: "ap-northeast-2", EventBusName: "bus"}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
}

func streamRelayEventForEventBridgeTest(at time.Time) StreamRelayEvent {
	return StreamRelayEvent{
		SchemaVersion:  StreamRelayEventSchemaVersion,
		StreamARN:      "arn:aws:dynamodb:ap-northeast-2:123456789012:table/events/stream/2026-05-26T05:06:07.000",
		ShardID:        "shard-1",
		SequenceNumber: "42",
		EventID:        "stream-event-1",
		EventName:      "INSERT",
		Record: OutboxRecord{
			SchemaVersion: OutboxRecordSchemaVersion,
			Event: TaskEvent{
				Seq:          1,
				TaskID:       "task-a",
				AssignmentID: "asn-000001",
				AgentID:      "jykim1",
				Type:         EventAssignmentQueued,
				State:        AssignmentQueued,
				At:           at,
			},
		},
	}
}
