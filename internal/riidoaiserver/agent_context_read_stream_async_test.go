package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type contextReadStreamSignalWriter struct {
	*httptest.ResponseRecorder
	written chan struct{}
}

func (w contextReadStreamSignalWriter) Write(p []byte) (int, error) {
	if strings.Contains(string(p), "agent_work_status_changed") {
		select {
		case w.written <- struct{}{}:
		default:
		}
	}
	return w.ResponseRecorder.Write(p)
}

func TestContextReadStreamSlowReadNeverBlocksTerminalOrWritesLate(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	r := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	w := contextReadStreamSignalWriter{httptest.NewRecorder(), make(chan struct{}, 1)}
	live := make(chan ClientStreamEvent, 1)
	started, release, finished, done := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	filter := func(_ context.Context, event ClientStreamEvent) (ClientStreamEvent, bool) {
		if event.EventType == AgentClientEventContextRead {
			close(started)
			<-release // Even a late adapter completion must not write after cancellation.
			defer close(finished)
		}
		return event, true
	}
	go func() { streamLiveAIAgentClientEvents(w, r, live, filter); close(done) }()
	live <- ClientStreamEvent{Seq: 1, EventType: AgentClientEventContextRead}
	<-started
	live <- ClientStreamEvent{Seq: 2, EventType: AgentClientEventWorkStatusChanged}
	select {
	case <-w.written:
	case <-time.After(time.Second):
		t.Fatal("terminal waited for metadata authorization")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("canceled subscriber waited for metadata worker")
	}
	close(release)
	<-finished
	if strings.Contains(w.Body.String(), "agent_context_read") || !strings.Contains(w.Body.String(), "id: 2\n") {
		t.Fatalf("cursor regressed or terminal lost: %s", w.Body.String())
	}
}
