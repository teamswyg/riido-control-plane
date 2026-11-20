package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func taskScopeRetryRequest() AIAgentTaskContextRequest {
	return AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: defaultAIAgentClientWorkspaceID, BearerToken: "synthetic-requester"}
}

func taskScopeRetryFixture(t *testing.T, status, failInitial, failPages int) (*AIAgentPrivateTaskContextClient, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	var initial, pages atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer synthetic-requester" && r.Header.Get("Authorization") != "Bearer ai-agent-token" {
			t.Error("requester bearer changed")
		}
		if strings.HasSuffix(r.URL.Path, "/workspace") {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "task-private", "team": map[string]any{"id": "team", "workspace": map[string]string{"id": defaultAIAgentClientWorkspaceID}}})
			return
		}
		var req struct{ Query string }
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Error("invalid query")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		root, value := "getComponentReadScope", readScopeCaptureInitial()
		var failed bool
		if strings.Contains(req.Query, "getComponentReadScopeBody") {
			failed = int(pages.Add(1)) <= failPages
			root, value = "getComponentReadScopeBody", readScopeCaptureBody(5, "world", nil, true)
		} else {
			failed = int(initial.Add(1)) <= failInitial
		}
		if failed {
			w.WriteHeader(status)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: value}})
	}))
	t.Cleanup(upstream.Close)
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, AssignmentReadScope: true, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return client, &initial, &pages
}
