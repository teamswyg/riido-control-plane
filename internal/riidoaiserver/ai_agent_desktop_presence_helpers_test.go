package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const desktopPresenceTestPath = "/v2/client/workspaces/workspace-beta/ai-agent/desktop-presence"

func readDesktopPresence(t *testing.T, server http.Handler, token string) AIAgentDesktopPresenceResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, desktopPresenceTestPath, nil)
	req.Header.Set(aiAgentTokenHeader, token)
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("desktop presence status=%d body=%s", resp.Code, resp.Body.String())
	}
	if strings.Contains(resp.Body.String(), "device_id") {
		t.Fatalf("desktop presence must not expose device identifiers: %s", resp.Body.String())
	}
	var body AIAgentDesktopPresenceResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("desktop presence json: %v", err)
	}
	return body
}

func connectDesktopForPresence(t *testing.T, server http.Handler, token, machineID string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-beta/devices/connect",
		strings.NewReader(`{"machine_id":"`+machineID+`"}`))
	req.Header.Set(aiAgentTokenHeader, token)
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("connect status=%d body=%s", resp.Code, resp.Body.String())
	}
}
