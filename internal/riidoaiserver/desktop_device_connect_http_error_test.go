package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPDesktopDeviceConnectValidatesMachineID(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "desktop-user",
		Token:       "desktop-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	req := httptest.NewRequest(
		http.MethodPost,
		"/v2/desktop/workspaces/workspace-beta/devices/connect",
		strings.NewReader(`{"machine_id":"   "}`),
	)
	req.Header.Set(aiAgentTokenHeader, "desktop-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest ||
		!strings.Contains(resp.Body.String(), "machine_id is required") {
		t.Fatalf("connect validation status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPDesktopDeviceConnectRequiresDaemonRuntime(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/v2/desktop/workspaces/workspace-beta/devices/connect",
		strings.NewReader(`{"machine_id":"machine-beta"}`),
	)
	resp := httptest.NewRecorder()
	NewServer(ServerConfig{}).Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusServiceUnavailable ||
		!strings.Contains(resp.Body.String(), "daemon runtime store is not configured") {
		t.Fatalf("connect missing runtime status=%d body=%s", resp.Code, resp.Body.String())
	}
}
