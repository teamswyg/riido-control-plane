package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIAgentReadPagesCurrentVisibleDirectoryWithoutRun(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	for i := range 52 {
		agent := store.agents["agent-public-openclaw"]
		agent.AgentID, agent.Name = fmt.Sprintf("page-agent-%02d", i), fmt.Sprintf("page agent %02d", i)
		store.agents[agent.AgentID] = agent
	}
	checks, prechecks := 0, 0
	allowed := true
	a := authorizationFunc(func(_ context.Context, _ string, req AuthorizationRequest) (AuthorizationResult, error) {
		checks++
		if !allowed {
			return AuthorizationResult{}, ErrAuthorizationForbidden
		}
		return AuthorizationResult{PrincipalID: "user-3", WorkspaceID: req.WorkspaceID}, nil
	})
	s := NewServer(ServerConfig{AIAgentClient: store, CLIAgentReadAuthorizer: a, AgentMentionPrecheck: func(context.Context, Assignment, string) ([]string, error) { prechecks++; return nil, nil }})
	seen := map[string]bool{}
	for i, suffix := range []string{"", "?cursor=1", "?cursor=9223372036854775807"} {
		w := cliAgentDirectoryRecorder(s, suffix)
		var page AgentDirectoryPage
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &page) != nil || page.Total != 53 || page.PolicyChecked || page.FetchedAt.IsZero() {
			t.Fatal("page projection invalid", w.Code)
		}
		if (i == 0 && (len(page.Agents) != 50 || page.NextCursor != "1")) || (i == 1 && (len(page.Agents) != 3 || page.NextCursor != "")) || (i == 2 && len(page.Agents) != 0) {
			t.Fatal("page bound changed", i, len(page.Agents), page.NextCursor)
		}
		for _, agent := range page.Agents {
			if seen[agent.AgentID] || agent.IsCurrentAgent {
				t.Fatal("duplicate or synthetic current agent")
			}
			seen[agent.AgentID] = true
		}
	}
	allowed = false
	if w := cliAgentDirectoryRecorder(s, "?cursor=1"); w.Code != http.StatusForbidden || strings.Contains(w.Body.String(), "page-agent") {
		t.Fatal("revoked current authority leaked page")
	}
	if checks != 4 || prechecks != 0 || len(seen) != 53 {
		t.Fatal("read reused authority or fabricated a run", checks, prechecks)
	}
}

func cliAgentDirectoryRecorder(s Server, suffix string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/v1/cli/workspaces/workspace-dev-riid/tasks/task-fixture/agents"+suffix, nil)
	r.Header.Set("Authorization", "Bearer "+CLIAgentReadTokenPrefix+strings.Repeat("A", 43))
	w := httptest.NewRecorder()
	s.handleCLIAgentDirectory(w, r)
	return w
}
