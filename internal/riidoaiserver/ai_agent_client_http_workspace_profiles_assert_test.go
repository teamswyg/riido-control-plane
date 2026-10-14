package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func assertWorkspaceAssignedProfileColor(t *testing.T, server http.Handler, taskID, wantColor string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/assigned-agent-profiles", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("assigned profiles after assign status=%d body=%s", resp.Code, resp.Body.String())
	}
	var next AssignedAgentProfileMapResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &next); err != nil {
		t.Fatalf("assigned profiles after assign json: %v", err)
	}
	if got := next.AssignedAgentProfiles[taskID]; got.TmpColor != wantColor {
		t.Fatalf("component keyed profile = %+v, all=%+v", got, next.AssignedAgentProfiles)
	}
}
