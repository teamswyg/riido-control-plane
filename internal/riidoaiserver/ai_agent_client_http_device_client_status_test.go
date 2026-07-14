package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientDevicesIncludesVersionGuidance(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1", Token: "user-token", Scopes: []string{"ai-agent:*"},
	}})
	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices", nil)
	req.Header.Set(aiAgentTokenHeader, "user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("devices status=%d body=%s", resp.Code, resp.Body.String())
	}
	var result DeviceRuntimeListResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Devices) == 0 || result.Devices[0].ClientStatus == nil {
		t.Fatalf("missing device client status: %s", resp.Body.String())
	}
	status := result.Devices[0].ClientStatus
	if status.MinimumDaemonVersion == "" || status.DownloadURL == "" {
		t.Fatalf("incomplete version guidance: %+v", status)
	}
}
