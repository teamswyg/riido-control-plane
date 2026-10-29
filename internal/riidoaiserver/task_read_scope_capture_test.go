package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTaskReadScopeActualHTTPBodyPagesReachAssignmentCLI(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer synthetic-requester" {
			t.Error("requester authorization missing")
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/workspace") {
			_, _ = w.Write([]byte(`{"id":"task-private","team":{"id":"team","workspace":{"id":"workspace-dev-riid"}}}`))
			return
		}
		if r.URL.Path != "/graphql" {
			_, _ = w.Write([]byte(`{"id":"task-private","title":"root"}`))
			return
		}
		var req struct {
			Query     string
			Variables map[string]any
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Fatal("invalid graphql request")
		}
		calls.Add(1)
		if strings.Contains(req.Query, "getComponentReadScopeBody") {
			if req.Variables["offset"] != float64(5) || req.Variables["version"] != "sha256:223ffd29e8e3e19af5e6b4c9419acdc8bb5f3a58c0bec336da5fa04a5a96e305" {
				t.Errorf("continuation variables=%v", req.Variables)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScopeBody": readScopeCaptureBody(5, "world", nil, true)}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": readScopeCaptureInitial()}})
	}))
	t.Cleanup(upstream.Close)
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, AssignmentReadScope: true})
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace-dev-riid", BearerToken: "synthetic-requester"})
	if err != nil || got.ReadScope == nil || calls.Load() != 2 {
		t.Fatalf("read scope absent or not paged: scope=%v calls=%d err=%v", got.ReadScope, calls.Load(), err)
	}
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	f.contexts.mu.Lock()
	delete(f.contexts.snapshots, f.leased.ID)
	f.contexts.mu.Unlock()
	f.service.config.AssignmentReadScope = true
	f.service.captureAssignmentContext(t.Context(), f.leased, &got, assignmentContextTrigger{Kind: AssignmentContextTriggerAssign})
	snapshot, found, err := f.contexts.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil || !found || snapshot.Context.ReadScope == nil {
		t.Fatalf("bundle capture missing: %v %v", found, err)
	}
	readScopeAssertBodyHTTP(t, f, snapshot)
}

func readScopeCaptureBody(offset int, content string, next any, complete bool) map[string]any {
	return map[string]any{"resourceId": "task-private", "part": "document", "status": "available", "source": "provider_json", "format": "html", "version": "sha256:223ffd29e8e3e19af5e6b4c9419acdc8bb5f3a58c0bec336da5fa04a5a96e305", "offset": offset, "content": content, "nextOffset": next, "complete": complete}
}

func readScopeCaptureInitial() map[string]any {
	return map[string]any{"schemaVersion": "riido-read-scope.v1", "rootId": "task-private", "teamId": "team", "status": "available", "selection": map[string]int{"relatedWorkLimit": 5, "relatedMeetingLimit": 5, "childrenLimit": 10}, "meetings": []any{}, "works": []any{map[string]any{"id": "task-private", "kind": "component", "teamId": "team", "reasons": []any{map[string]string{"kind": "root"}}, "details": map[string]any{"id": "task-private", "title": "root", "componentType": "TASK", "keyNumber": "12", "updatedAt": "2026-10-28T00:00:00Z"}, "body": readScopeCaptureBody(0, "hello", 5, false), "comments": map[string]any{"resourceId": "task-private", "resourceKind": "component", "version": "comments-v1", "status": "empty", "page": 1, "limit": 20, "nextPage": nil, "complete": true, "items": []any{}}, "links": []any{}, "github": []any{}, "repositories": []any{}, "figma": nil, "priorities": []any{}, "checklists": []any{}, "attachments": []any{}, "sprints": []any{}, "origin": nil, "partsStatus": []any{}}}}
}
