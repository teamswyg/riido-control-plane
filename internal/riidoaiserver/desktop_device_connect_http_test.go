package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPDesktopDeviceConnectAddsWorkspace(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "desktop-user",
		Token:       "desktop-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	req := httptest.NewRequest(
		http.MethodPost,
		"/v2/desktop/workspaces/workspace-beta/devices/connect",
		strings.NewReader(`{"machine_id":"machine-beta"}`),
	)
	req.Header.Set(aiAgentTokenHeader, "desktop-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("connect status=%d body=%s", resp.Code, resp.Body.String())
	}
	var body struct {
		SchemaVersion         string   `json:"schema_version"`
		DeviceID              string   `json:"device_id"`
		ConnectedWorkspaceIDs []string `json:"connected_workspace_ids"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("connect json: %v", err)
	}
	if body.SchemaVersion != SchemaVersion || body.DeviceID == "" ||
		!containsString(body.ConnectedWorkspaceIDs, "workspace-beta") {
		t.Fatalf("connect response = %+v", body)
	}
}
