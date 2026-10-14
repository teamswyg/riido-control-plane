package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientWorkspaceAssignedAgentProfilesReconcileError(t *testing.T) {
	assignment := NewStore()
	t.Cleanup(assignment.Close)

	server := newWorkspaceProfilesErrorTestServer(
		t,
		[]string{"ai-agent:*"},
		workspaceProfilesErrorStore{
			DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
			reconcileErr:                  errors.New("projection reconcile failed"),
		},
		assignment,
	)

	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/assigned-agent-profiles", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadGateway {
		t.Fatalf("assigned profiles status=%d want=%d body=%s", resp.Code, http.StatusBadGateway, resp.Body.String())
	}
}
