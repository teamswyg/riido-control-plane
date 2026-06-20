package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentDevicesAndEditability(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	devicesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/devices", nil)
	devicesReq.Header.Set("Authorization", "Bearer user-token")
	devicesResp := httptest.NewRecorder()
	server.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("devices status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("devices json: %v", err)
	}
	if len(devices.Devices) != 2 {
		t.Fatalf("devices = %+v", devices)
	}
	ownedDevice, ok := findDevice(devices.Devices, "device-dev-macbook")
	if !ok || len(ownedDevice.Runtimes) != 3 {
		t.Fatalf("owned device = %+v, ok=%v", ownedDevice, ok)
	}
	sharedDevice, ok := findDevice(devices.Devices, "device-shared-studio")
	if !ok || len(sharedDevice.Runtimes) != 1 || sharedDevice.Runtimes[0].RuntimeID != "runtime-openclaw-shared" {
		t.Fatalf("shared public-agent device = %+v, ok=%v", sharedDevice, ok)
	}
	cursorRuntime := ownedDevice.Runtimes[2]
	if cursorRuntime.RuntimeID != "runtime-cursor-dev" || len(cursorRuntime.Models) != 2 || cursorRuntime.Models[0].ModelID != "cursor-auto" || !cursorRuntime.Models[0].IsDefault {
		t.Fatalf("cursor runtime models = %+v", cursorRuntime)
	}

	editReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-owned-codex/editability", nil)
	editReq.Header.Set("Authorization", "Bearer user-token")
	editResp := httptest.NewRecorder()
	server.ServeHTTP(editResp, editReq)
	if editResp.Code != http.StatusOK {
		t.Fatalf("editability status=%d body=%s", editResp.Code, editResp.Body.String())
	}
	var edit AgentEditabilityResponse
	if err := json.Unmarshal(editResp.Body.Bytes(), &edit); err != nil {
		t.Fatalf("editability json: %v", err)
	}
	if edit.Editability != AgentEditabilityBlockedAssignedTasks || edit.AssignedTaskCount != 1 || edit.Reason == "" {
		t.Fatalf("editability = %+v", edit)
	}
}
