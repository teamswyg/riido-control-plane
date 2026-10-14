package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientAssignableAgentsMatchAssignmentRuntimeValidation(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:component-a:read"},
	}})
	assignablePath := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID +
		"/ai-agent/tasks/component-a/assignable-agents"
	var assignable AgentClientListResponse
	aiAgentSmokeDecode(t, aiAgentSmokeRequest(
		t, server, http.MethodGet, assignablePath, "user-token", "", http.StatusOK,
	), &assignable)
	ids := aiAgentIDs(assignable.Agents)
	if containsString(ids, "agent-owned-claude") {
		t.Fatalf("assignable agents included unavailable runtime: %v", ids)
	}
	if !containsString(ids, "agent-owned-codex") {
		t.Fatalf("assignable agents omitted available runtime: %v", ids)
	}

	assignmentPath := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID +
		"/ai-agent/tasks/component-a/agent-assignments"
	aiAgentSmokeRequest(t, server, http.MethodPost, assignmentPath, "user-token",
		`{"agent_id":"agent-owned-claude"}`, http.StatusBadRequest)
}
