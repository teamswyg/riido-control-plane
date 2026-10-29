package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskReadScopeNormalTwentyFiveWorksAndFiveMeetingsCompleteWithinQuota(t *testing.T) {
	initial, meetings := readScopeWholeFixture()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/workspace") {
			_, _ = w.Write([]byte(`{"id":"task-private","team":{"id":"team","workspace":{"id":"workspace-dev-riid"}}}`))
			return
		}
		var req struct {
			Query     string
			Variables map[string]any
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Error("invalid request")
		}
		root, value := "getComponentReadScope", any(initial)
		if strings.Contains(req.Query, "getMeetingReadScope(") {
			root = "getMeetingReadScope"
			value = meetings[req.Variables["meetingId"].(string)]
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: value}})
	}))
	t.Cleanup(server.Close)
	c, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL, AssignmentReadScope: true})
	got, err := c.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace-dev-riid", BearerToken: "synthetic"})
	if err != nil || got.ReadScope == nil || len(got.ReadScope.Resources) != 30 {
		t.Fatalf("scope capture err=%v", err)
	}
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	f.contexts.mu.Lock()
	delete(f.contexts.snapshots, f.leased.ID)
	f.contexts.mu.Unlock()
	f.service.config.AssignmentReadScope = true
	f.service.captureAssignmentContext(t.Context(), f.leased, &got, assignmentContextTrigger{Kind: AssignmentContextTriggerAssign})
	snapshot, found, err := f.contexts.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil || !found || snapshot.Truncated {
		t.Fatalf("normal fixture not complete: %v %v truncated=%v", found, err, snapshot.Truncated)
	}
	assertReadScopeWholeInline(t, f, got.ReadScope.Resources)
	if len(snapshot.Content) != 120 || len(snapshot.ContentBlocks) > 10 {
		t.Fatalf("packing logical=%d physical=%d", len(snapshot.Content), len(snapshot.ContentBlocks))
	}
}
