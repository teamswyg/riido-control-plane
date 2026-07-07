package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientStopTaskAgentAssignmentRejectsUnscopedRouteAndBadJSON(t *testing.T) {
	const token = "user-token"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*", "task:task-new:stop"},
	}})

	path := "/v1/client/ai-agent/tasks/task-new/agent-assignments/agent-public-openclaw/stop"
	aiAgentSmokeRequest(t, server, http.MethodPost, path, token, `{}`, http.StatusNotFound)

	path = "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-new/agent-assignments/agent-public-openclaw/stop"
	aiAgentSmokeRequest(t, server, http.MethodPost, path, token, `{`, http.StatusBadRequest)
}
