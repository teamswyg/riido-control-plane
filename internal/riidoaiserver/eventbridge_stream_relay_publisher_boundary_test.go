package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestEventBridgeStreamRelayPublisherRejectsInvalidEventsBeforeRequest(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		atomic.AddInt32(&calls, 1)
	}))
	defer server.Close()
	publisher := newEventBridgeBoundaryPublisher(t, server.URL)
	defer publisher.Close()
	valid := streamRelayEventForEventBridgeTest(time.Date(2026, 5, 26, 5, 6, 7, 0, time.UTC))
	cases := []struct {
		name  string
		event StreamRelayEvent
		want  string
	}{
		{"event-schema", withStreamRelayEventSchema(valid, "old"), "stream relay event"},
		{"outbox-schema", withOutboxRecordSchema(valid, "old"), "stream relay outbox"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := publisher.PublishStreamRelayEvent(context.Background(), tc.event)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("PublishStreamRelayEvent error = %v, want %q", err, tc.want)
			}
		})
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("unexpected EventBridge requests = %d", got)
	}
}

func TestEventBridgeStreamRelayPublisherNilAndClosedBoundaries(t *testing.T) {
	var nilPublisher *EventBridgeStreamRelayPublisher
	if err := nilPublisher.PublishStreamRelayEvent(context.Background(), StreamRelayEvent{}); err != nil {
		t.Fatalf("nil publish err=%v", err)
	}
	if err := nilPublisher.Close(); err != nil {
		t.Fatalf("nil close err=%v", err)
	}
	publisher := newEventBridgeBoundaryPublisher(t, "http://127.0.0.1:9")
	if err := publisher.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	err := publisher.PublishStreamRelayEvent(context.Background(), StreamRelayEvent{})
	if err == nil || !strings.Contains(err.Error(), "publisher closed") {
		t.Fatalf("closed publish err=%v", err)
	}
}

func newEventBridgeBoundaryPublisher(t *testing.T, endpoint string) *EventBridgeStreamRelayPublisher {
	t.Helper()
	publisher, err := NewEventBridgeStreamRelayPublisher(EventBridgeStreamRelayPublisherConfig{
		Region:              "ap-northeast-2",
		EventBusName:        "riido-ai-server-events",
		Endpoint:            endpoint,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKID", ""),
	})
	if err != nil {
		t.Fatalf("NewEventBridgeStreamRelayPublisher: %v", err)
	}
	return publisher
}
