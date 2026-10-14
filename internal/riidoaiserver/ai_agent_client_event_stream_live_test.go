package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAIAgentClientEventStreamWritesLiveEvent(t *testing.T) {
	live := make(chan ClientStreamEvent, 1)
	live <- ClientStreamEvent{
		Seq:       42,
		EventType: AgentClientEventThreadProgress,
		Payload: AgentThreadProgressEvent{
			EventType: AgentClientEventThreadProgress,
			TaskID:    "task-live",
			ThreadID:  "thread-live",
		},
	}
	close(live)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events", nil)

	streamLiveAIAgentClientEvents(rec, req, live)

	body := rec.Body.String()
	if !strings.Contains(body, "id: 42\n") ||
		!strings.Contains(body, "event: agent_thread_progress\n") ||
		!strings.Contains(body, `"thread_id":"thread-live"`) {
		t.Fatalf("live SSE body = %q", body)
	}
	if !rec.Flushed {
		t.Fatalf("live SSE did not flush after event")
	}
}

func TestAIAgentClientEventStreamStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events", nil).WithContext(ctx)
	done := make(chan struct{})
	go func() {
		streamLiveAIAgentClientEvents(httptest.NewRecorder(), req, make(chan ClientStreamEvent))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("stream did not stop after context cancellation")
	}
}
