package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadSSEParsesProgressAndInvalidPayloads(t *testing.T) {
	t.Parallel()
	raw := "event: agent_thread_progress\n" +
		`data: {"task_id":"task","thread_id":"thread","conversation_id":"conv",` +
		`"assignment_id":"asn","run_id":"run","work_status":"running",` +
		`"assignment_state":"running","lines":[{"seq":1,"message":"thinking"}]}` +
		"\n\n" +
		"event: malformed\n" +
		"data: not-json\n\n"
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(raw))}
	events := readSSE(context.Background(), resp)
	if len(events) != 2 {
		t.Fatalf("events len = %d", len(events))
	}
	if events[0].Event != "agent_thread_progress" || events[0].LineCount != 1 {
		t.Fatalf("progress event not parsed: %+v", events[0])
	}
	if events[1].Event != "malformed" || events[1].ThreadID != "" {
		t.Fatalf("malformed event should remain redacted/empty: %+v", events[1])
	}
}

func TestAppendSSEIgnoresEmptyPayload(t *testing.T) {
	t.Parallel()
	events := []sseEventSummary{{Event: "existing"}}
	got := appendSSE(events, "ignored", "")
	if len(got) != 1 || got[0].Event != "existing" {
		t.Fatalf("empty payload should not append: %+v", got)
	}
}

func TestReadSSEStopsOnCancelledContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("event: x\n"))}
	if events := readSSE(ctx, resp); len(events) != 0 {
		t.Fatalf("cancelled context should return no events: %+v", events)
	}
}
