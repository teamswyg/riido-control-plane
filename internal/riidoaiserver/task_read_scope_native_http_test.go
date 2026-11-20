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
	var calls, intakeAttempts int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/workspace") {
			t.Error("native read must not call legacy workspace locator")
			http.NotFound(w, r)
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
		if strings.Contains(req.Query, "getComponentWorkspace(") {
			if r.Header.Get("Authorization") != "Bearer human" || req.Variables["id"] != "task-private" {
				t.Error("native locator lost requester or identity")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentWorkspace": nativeTestLocation()}})
			return
		}
		if r.Header.Get("Authorization") != "Bearer human" || req.Variables["workspaceId"] != "workspace" || req.Variables["locatorVersion"] != float64(2) || !strings.Contains(req.Query, "workspaceId:$workspaceId,locatorVersion:$locatorVersion") {
			t.Errorf("native request lost authority: %+v", req)
		}
		root, data := "getComponentReadScope", any(capture.Scope)
		if strings.Contains(req.Query, "getComponentReadScope(componentId:") && req.Variables["teamId"] != "team" {
			t.Error("producer requires the current team resolved by the locator")
			_, _ = w.Write([]byte(`{"errors":[{"extensions":{"code":"permission_denied"}}]}`))
			return
		}
		if strings.Contains(req.Query, "getIntakeReadScopeBody") {
			intakeAttempts++
			if intakeAttempts == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
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
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL + "/rest-not-routed", GraphQLURL: upstream.URL + "/graphql", LocatorVersion: 2})
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace", BearerToken: "human"})
	if err != nil || calls != 5 || intakeAttempts != 2 || got.Component.Kind != "intake" || got.TeamID != "team" || got.ReadScope.TeamID != "team" || got.Hierarchy.Project.ID != "project" {
		t.Fatalf("native read failed: calls=%d context=%+v err=%v", calls, got, err)
	}
	composed, err := composeAssignRequestWithTaskContextResult("task-private", "task-private", AssignRequest{Prompt: "Inspect only"}, got)
	if err != nil || composed.Request.TeamID != "team" || !strings.Contains(composed.Request.Prompt, "hello") || !strings.Contains(composed.Request.Prompt, "current project") || strings.Contains(composed.Request.Prompt, taskContextUnavailableDocumentContent) {
		t.Fatalf("native context did not reach runtime prompt: %+v %v", composed, err)
	}
}
