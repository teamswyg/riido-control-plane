package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDeviceDaemonDetailIncludesAppVersion(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-app-version",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})
	enroll := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", strings.NewReader(`{"display_name":"Version Mac","platform":"darwin"}`))
	enroll.Header.Set(aiAgentTokenHeader, "user-token")
	enrollResp := httptest.NewRecorder()
	server.ServeHTTP(enrollResp, enroll)
	if enrollResp.Code != http.StatusCreated {
		t.Fatalf("enroll status=%d body=%s", enrollResp.Code, enrollResp.Body.String())
	}
	var enrollment EnrollDeviceResponse
	if err := json.Unmarshal(enrollResp.Body.Bytes(), &enrollment); err != nil {
		t.Fatalf("enroll json: %v", err)
	}

	body := `{"daemon_id":"daemon-version","app_version":"riido-daemon v0.0.39","runtimes":[{"runtime_id":"version-device:codex","kind":"codex"}]}`
	sync := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(body))
	sync.Header.Set(deviceIDHeader, enrollment.DeviceID)
	sync.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	syncResp := httptest.NewRecorder()
	server.ServeHTTP(syncResp, sync)
	if syncResp.Code != http.StatusAccepted {
		t.Fatalf("runtime sync status=%d body=%s", syncResp.Code, syncResp.Body.String())
	}

	detail := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/"+enrollment.DeviceID+"/daemon", nil)
	detail.Header.Set(aiAgentTokenHeader, "user-token")
	detailResp := httptest.NewRecorder()
	server.ServeHTTP(detailResp, detail)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("daemon detail status=%d body=%s", detailResp.Code, detailResp.Body.String())
	}
	var daemon DeviceDaemonDetailResponse
	if err := json.Unmarshal(detailResp.Body.Bytes(), &daemon); err != nil {
		t.Fatalf("daemon detail json: %v", err)
	}
	if daemon.Daemon.AppVersion != "riido-daemon v0.0.39" {
		t.Fatalf("daemon app_version=%q", daemon.Daemon.AppVersion)
	}
}
