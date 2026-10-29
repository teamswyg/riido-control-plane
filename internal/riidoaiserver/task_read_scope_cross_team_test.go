package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskReadScopeCapturesAuthorizedLeafTeams(t *testing.T) {
	scope, meeting, last := readScopeCrossTeamFixture()
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
			t.Error("invalid query")
			return
		}
		root, value, team := "getComponentReadScope", any(scope), "team"
		if strings.Contains(req.Query, "getComponentReadScopeBody(") {
			root, value, team = "getComponentReadScopeBody", last, "leaf-team"
		}
		if strings.Contains(req.Query, "getMeetingReadScope(") {
			root, value, team = "getMeetingReadScope", meeting, "leaf-team"
		}
		if req.Variables["teamId"] != team {
			t.Errorf("wrong leaf team: %v", req.Variables["teamId"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: value}})
	}))
	t.Cleanup(server.Close)
	c, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL, AssignmentReadScope: true})
	got, err := c.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace-dev-riid", BearerToken: "synthetic"})
	if err != nil || got.ReadScope == nil || len(got.ReadScope.Resources) != 3 {
		t.Fatalf("authorized leaf capture: %v", err)
	}
	snapshot := AssignmentContextSnapshot{Context: got, Content: got.readScopeCapture.Manifest}
	if err := validateAssignmentReadScopeManifest(snapshot); err != nil {
		t.Fatal(err)
	}
	for _, resource := range got.ReadScope.Resources[1:] {
		if resource.TeamID != "leaf-team" {
			t.Fatal("leaf team lost")
		}
	}
	snapshot.Context.ReadScope.Resources[0].TeamID = "leaf-team"
	if validateAssignmentReadScopeManifest(snapshot) == nil {
		t.Fatal("root team mismatch accepted")
	}
}
