package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentAdminCreateUsesAuthorizedWorkspaceRuntime(t *testing.T) {
	adminServer := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "admin-1",
		Token:       "admin-token",
		Scopes:      []string{"ai-agent:*"},
		Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
	}})

	devicesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/devices", nil)
	devicesReq.Header.Set("Authorization", "Bearer admin-token")
	devicesResp := httptest.NewRecorder()
	adminServer.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("admin devices status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("admin devices json: %v", err)
	}
	if len(devices.Devices) != 2 {
		t.Fatalf("admin devices = %+v", devices.Devices)
	}
	ownedDevice, ok := findDevice(devices.Devices, "device-dev-macbook")
	if !ok || ownedDevice.OwnerPrincipalID != "user-1" || len(ownedDevice.Runtimes) != 3 {
		t.Fatalf("admin owned device = %+v, ok=%v", ownedDevice, ok)
	}
	sharedDevice, ok := findDevice(devices.Devices, "device-shared-studio")
	if !ok || sharedDevice.OwnerPrincipalID != "user-2" || len(sharedDevice.Runtimes) != 2 {
		t.Fatalf("admin shared device = %+v, ok=%v", sharedDevice, ok)
	}

	createBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:       "관리자 생성 에이전트",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
	})
	if err != nil {
		t.Fatalf("marshal admin create body: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents", strings.NewReader(string(createBody)))
	createReq.Header.Set("Authorization", "Bearer admin-token")
	createResp := httptest.NewRecorder()
	adminServer.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("admin create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var created AgentClientRecordResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("admin create json: %v", err)
	}
	if created.Agent.OwnerPrincipalID != "admin-1" ||
		created.Agent.RuntimeID != "runtime-cursor-dev" ||
		created.Agent.RuntimeKind != RuntimeKindCursor ||
		!created.Agent.IsOwnedByViewer {
		t.Fatalf("admin created agent = %+v", created.Agent)
	}

	assertNonAdminRuntimeCreateDenied(t, createBody)
}
