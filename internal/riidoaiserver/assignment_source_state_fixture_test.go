package riidoaiserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

const sourceStateTestKey = "source-state-test-key-0123456789abcdef"

type sourceStateHandler func(w http.ResponseWriter, body map[string]any)

// newSourceStateTestServer stands in for the monolith lookup and counts calls.
func newSourceStateTestServer(t *testing.T, handler sourceStateHandler) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	calls := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		if r.Method != http.MethodPost || r.URL.Path != "/internal/agent-source-states" || r.Header.Get(AgentSourceStateKeyHeader) != sourceStateTestKey || json.Unmarshal(raw, &body) != nil {
			http.Error(w, "rejected", http.StatusUnauthorized)
			return
		}
		handler(w, body)
	}))
	t.Cleanup(server.Close)
	return server, calls
}

func echoSourceStates(states map[string]string) sourceStateHandler {
	return func(w http.ResponseWriter, body map[string]any) {
		out := []map[string]any{}
		for _, item := range body["sources"].([]any) {
			source := item.(map[string]any)
			out = append(out, map[string]any{"kind": source["kind"], "id": source["id"], "team_id": source["team_id"], "state": states[source["id"].(string)]})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"states": out})
	}
}

// sourceStateTestServer stores a previous capture with three related sources
// and returns a follow-up capture in the same conversation without them.
func sourceStateTestServer(t *testing.T, server *httptest.Server, extra ...string) (Server, AssignmentContextSnapshot) {
	t.Helper()
	store := NewMemoryAssignmentContextStore()
	if len(extra) == 0 {
		extra = []string{"gone", "moved", "secret"}
	}
	previous, pages := sourceStateTestBundle(t, "previous", extra...)
	if err := store.SaveAssignmentContentBundle(t.Context(), previous, pages); err != nil {
		t.Fatal(err)
	}
	current, _ := sourceStateTestBundle(t, "current", "moved-in")
	current.PreviousAssignmentID, current.PreviousThreadID = previous.AssignmentID, previous.ThreadID
	config := AgentSourceStateConfig{BaseURL: server.URL, Key: sourceStateTestKey, HTTPClient: server.Client()}
	return Server{config: ServerConfig{AssignmentContext: store, AgentSourceState: config}}, current
}
