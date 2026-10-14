package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientExternalWorkspaceHidesDevelopmentSeedDevices(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "external-admin",
		Token:       "external-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
		Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
	}, {
		PrincipalID: "other-user",
		Token:       "other-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})

	enrollReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-real/devices/enroll", strings.NewReader(`{"display_name":"Real Mac","platform":"darwin","app_version":"0.0.0"}`))
	enrollReq.Header.Set(aiAgentTokenHeader, "external-token")
	enrollResp := httptest.NewRecorder()
	server.ServeHTTP(enrollResp, enrollReq)
	if enrollResp.Code != http.StatusCreated {
		t.Fatalf("enroll status=%d body=%s", enrollResp.Code, enrollResp.Body.String())
	}
	var enrollment EnrollDeviceResponse
	if err := json.Unmarshal(enrollResp.Body.Bytes(), &enrollment); err != nil {
		t.Fatalf("enroll json: %v", err)
	}

	otherEnrollReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-other/devices/enroll", strings.NewReader(`{"display_name":"Other Mac","platform":"darwin","app_version":"0.0.0"}`))
	otherEnrollReq.Header.Set(aiAgentTokenHeader, "other-token")
	otherEnrollResp := httptest.NewRecorder()
	server.ServeHTTP(otherEnrollResp, otherEnrollReq)
	if otherEnrollResp.Code != http.StatusCreated {
		t.Fatalf("other enroll status=%d body=%s", otherEnrollResp.Code, otherEnrollResp.Body.String())
	}
	var otherEnrollment EnrollDeviceResponse
	if err := json.Unmarshal(otherEnrollResp.Body.Bytes(), &otherEnrollment); err != nil {
		t.Fatalf("other enroll json: %v", err)
	}

	syncBody := `{"daemon_id":"daemon-real","runtimes":[{"runtime_id":"agentd-real:codex","kind":"codex","availability":"online","detection_state":"detected","models":[{"model_id":"gpt-5.5","label":"gpt-5.5","is_default":true}]}]}`
	syncReq := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(syncBody))
	syncReq.Header.Set(deviceIDHeader, enrollment.DeviceID)
	syncReq.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	syncResp := httptest.NewRecorder()
	server.ServeHTTP(syncResp, syncReq)
	if syncResp.Code != http.StatusAccepted {
		t.Fatalf("runtime sync status=%d body=%s", syncResp.Code, syncResp.Body.String())
	}

	devicesReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-real/ai-agent/devices", nil)
	devicesReq.Header.Set(aiAgentTokenHeader, "external-token")
	devicesResp := httptest.NewRecorder()
	server.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("devices status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("devices json: %v", err)
	}
	if len(devices.Devices) != 1 {
		t.Fatalf("external workspace devices = %+v, want only enrolled device", devices.Devices)
	}
	if _, ok := findDevice(devices.Devices, "device-dev-macbook"); ok {
		t.Fatalf("external workspace leaked device-dev-macbook: %+v", devices.Devices)
	}
	if _, ok := findDevice(devices.Devices, "device-shared-studio"); ok {
		t.Fatalf("external workspace leaked device-shared-studio: %+v", devices.Devices)
	}
	if _, ok := findDevice(devices.Devices, otherEnrollment.DeviceID); ok {
		t.Fatalf("external workspace leaked other enrolled device: %+v", devices.Devices)
	}
	realDevice, ok := findDevice(devices.Devices, enrollment.DeviceID)
	if !ok || realDevice.OwnerPrincipalID != "external-admin" || len(realDevice.Runtimes) != 1 || realDevice.Runtimes[0].RuntimeID != "agentd-real:codex" {
		t.Fatalf("external workspace real device = %+v, ok=%v", realDevice, ok)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-real/ai-agent/bootstrap?client_kind=web", nil)
	bootstrapReq.Header.Set(aiAgentTokenHeader, "external-token")
	bootstrapResp := httptest.NewRecorder()
	server.ServeHTTP(bootstrapResp, bootstrapReq)
	if bootstrapResp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResp.Code, bootstrapResp.Body.String())
	}
	var bootstrap ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapResp.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("bootstrap json: %v", err)
	}
	if len(bootstrap.Devices) != 1 || bootstrap.Devices[0].DeviceID != enrollment.DeviceID {
		t.Fatalf("external workspace bootstrap devices = %+v", bootstrap.Devices)
	}
}
