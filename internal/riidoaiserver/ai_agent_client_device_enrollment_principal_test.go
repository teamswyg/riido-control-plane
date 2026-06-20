package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func (f *deviceEnrollmentHTTPFixture) verifyEnrolledDevicePrincipal() {
	t := f.t
	devicesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/devices", nil)
	devicesReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	devicesResp := httptest.NewRecorder()
	f.server.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("devices status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("devices json: %v", err)
	}
	if device, ok := findDevice(devices.Devices, f.enrollment.DeviceID); !ok || device.OwnerPrincipalID != "user-1" {
		t.Fatalf("enrolled device missing from devices response: %+v", devices.Devices)
	}
	devicePrincipal, err := f.aiAgentStore.AuthorizeDeviceCredential(context.Background(), f.enrollment.DeviceID, f.enrollment.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll})
	if err != nil {
		t.Fatalf("AuthorizeDeviceCredential: %v", err)
	}
	if devicePrincipal.PrincipalID != "user-1" || devicePrincipal.WorkspaceID != "workspace-alpha" {
		t.Fatalf("device principal = %+v", devicePrincipal)
	}
}
