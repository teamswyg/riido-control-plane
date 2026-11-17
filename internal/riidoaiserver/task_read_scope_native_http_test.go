package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskReadScopeNativeHTTPPropagatesWorkspaceOnContinuations(t *testing.T) {
	capture := nativeTestCapture(t, "intake")
	var calls int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/workspace") {
			_, _ = w.Write([]byte(`{"id":"task-private","team":{"id":"team","workspace":{"id":"workspace"}}}`))
			return
		}
		var req struct {
			Query     string
			Variables map[string]any
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		calls++
		if r.Header.Get("Authorization") != "Bearer human" || req.Variables["workspaceId"] != "workspace" || req.Variables["locatorVersion"] != float64(2) || !strings.Contains(req.Query, "workspaceId:$workspaceId,locatorVersion:$locatorVersion") {
			t.Errorf("native request lost authority: %+v", req)
		}
		root, data := "getComponentReadScope", any(capture.Scope)
		if strings.Contains(req.Query, "getIntakeReadScopeBody") {
			root, data = "getIntakeReadScopeBody", capture.BodyPages[0].Page
		}
		if strings.Contains(req.Query, "getComponentReadScopeBody") {
			root, data = "getComponentReadScopeBody", capture.BodyPages[1].Page
			if req.Variables["teamId"] != "" || strings.Contains(req.Query, "$teamId:String!") {
				t.Error("central parent requires nullable team")
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: data}})
	}))
	defer upstream.Close()
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2})
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace", BearerToken: "human"})
	if err != nil || calls != 3 || got.Component.Kind != "intake" {
		t.Fatalf("native read failed: calls=%d context=%+v err=%v", calls, got, err)
	}
}

func TestTaskReadScopeNativeNoLegacyFallback(t *testing.T) {
	for _, response := range []string{`{"errors":[{"extensions":{"code":"permission_denied"}}]}`, `{"data":{"getComponentReadScope":{"schemaVersion":"riido-read-scope.v1"}}}`} {
		calls := 0
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; _, _ = w.Write([]byte(response)) }))
		client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2})
		_, err := client.getTaskReadScope(t.Context(), AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace"}, "team")
		upstream.Close()
		if err == nil || calls != 1 {
			t.Fatalf("legacy fallback: calls=%d err=%v", calls, err)
		}
	}
}
