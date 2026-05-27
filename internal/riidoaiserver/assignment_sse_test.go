package riidoaiserver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStoreSubscribeTaskReceivesHistoryFanoutAndCancel(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	history, events, cancel, err := store.SubscribeTask(ctx, "task-a")
	if err != nil {
		t.Fatalf("SubscribeTask: %v", err)
	}
	if len(history) != 1 || history[0].Type != EventAssignmentQueued {
		t.Fatalf("history = %+v", history)
	}
	metrics, err := store.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics subscribed: %v", err)
	}
	if metrics.SSESubscribers != 1 {
		t.Fatalf("SSESubscribers subscribed = %d, want 1", metrics.SSESubscribers)
	}

	if _, err := store.RecordAgentEvent(ctx, "agent-a", AgentEventRequest{
		AssignmentID: assignment.ID,
		EventType:    EventRiidoLog,
		Message:      "stream me",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}
	select {
	case event := <-events:
		if event.Type != EventRiidoLog || event.Message != "stream me" {
			t.Fatalf("fanout event = %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for fanout event")
	}

	cancel()
	metrics, err = store.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics cancelled: %v", err)
	}
	if metrics.SSESubscribers != 0 {
		t.Fatalf("SSESubscribers cancelled = %d, want 0", metrics.SSESubscribers)
	}
	select {
	case _, ok := <-events:
		if ok {
			t.Fatal("events channel still open after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for subscriber channel close")
	}
}

func TestHTTPAssignmentSSEReplaysHistory(t *testing.T) {
	store := NewStore()
	defer store.Close()
	if _, err := store.AssignTask(context.Background(), "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	server := NewServer(ServerConfig{Assignment: store, Authorizer: assignmentHTTPAuthorizer(t, []string{"component-task:task-a:events:read"})}).Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/component-tasks/task-a/events?replay=1", nil)
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("sse replay status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("Content-Type = %q", got)
	}
	message := resp.Body.String()
	if !strings.Contains(message, "id: 1\n") || !strings.Contains(message, "event: assignment_queued\n") {
		t.Fatalf("sse replay message = %q", message)
	}
	var event TaskEvent
	if err := json.Unmarshal(sseData(t, message), &event); err != nil {
		t.Fatalf("sse data json: %v", err)
	}
	if event.TaskID != "task-a" || event.Type != EventAssignmentQueued {
		t.Fatalf("sse event = %+v", event)
	}
}

func TestHTTPAssignmentSSEStreamsNewEvents(t *testing.T) {
	store := NewStore()
	defer store.Close()
	assignment, err := store.AssignTask(context.Background(), "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	server := httptest.NewServer(NewServer(ServerConfig{Assignment: store, Authorizer: assignmentHTTPAuthorizer(t, []string{"component-task:task-a:events:read"})}).Handler())
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/v1/component-tasks/task-a/events", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("stream request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("stream status=%d body=%s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	history := readSSEMessage(t, reader)
	if !strings.Contains(history, "event: assignment_queued\n") {
		t.Fatalf("history sse = %q", history)
	}

	if _, err := store.RecordAgentEvent(context.Background(), "agent-a", AgentEventRequest{
		AssignmentID: assignment.ID,
		EventType:    EventRiidoLog,
		Message:      "hello stream",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}
	live := readSSEMessage(t, reader)
	if !strings.Contains(live, "event: riido_log\n") {
		t.Fatalf("live sse = %q", live)
	}
	var event TaskEvent
	if err := json.Unmarshal(sseData(t, live), &event); err != nil {
		t.Fatalf("live sse data json: %v", err)
	}
	if event.Message != "hello stream" || event.AssignmentID != assignment.ID {
		t.Fatalf("live event = %+v", event)
	}
}

func TestHTTPAssignmentSSERequiresScopedAuthorization(t *testing.T) {
	store := NewStore()
	defer store.Close()
	server := NewServer(ServerConfig{Assignment: store, Authorizer: assignmentHTTPAuthorizer(t, []string{"component-task:task-b:events:read"})}).Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/component-tasks/task-a/events?replay=1", nil)
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("other task sse status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPAssignmentSSEStoreNotConfiguredFailsClosed(t *testing.T) {
	server := NewServer(ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"component-task:task-a:events:read"})}).Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/component-tasks/task-a/events?replay=1", nil)
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured sse status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func readSSEMessage(t *testing.T, reader *bufio.Reader) string {
	t.Helper()
	type result struct {
		message string
		err     error
	}
	ch := make(chan result, 1)
	go func() {
		var b strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				ch <- result{message: b.String(), err: err}
				return
			}
			b.WriteString(line)
			if line == "\n" || line == "\r\n" {
				ch <- result{message: b.String()}
				return
			}
		}
	}()
	select {
	case res := <-ch:
		if res.err != nil {
			t.Fatalf("read SSE message: %v message=%q", res.err, res.message)
		}
		return res.message
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for SSE message")
		return ""
	}
}

func sseData(t *testing.T, message string) []byte {
	t.Helper()
	for _, line := range strings.Split(message, "\n") {
		if strings.HasPrefix(line, "data: ") {
			return []byte(strings.TrimPrefix(line, "data: "))
		}
	}
	t.Fatalf("SSE message missing data line: %q", message)
	return nil
}
