package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func (f *deviceEnrollmentHTTPFixture) verifyRuntimeMovedToReplacementDevice(replacement EnrollDeviceResponse) {
	t := f.t
	afterMoveDevicesReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices", nil)
	afterMoveDevicesReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	afterMoveDevicesResp := httptest.NewRecorder()
	f.server.ServeHTTP(afterMoveDevicesResp, afterMoveDevicesReq)
	if afterMoveDevicesResp.Code != http.StatusOK {
		t.Fatalf("after move devices status=%d body=%s", afterMoveDevicesResp.Code, afterMoveDevicesResp.Body.String())
	}
	var afterMoveDevices DeviceRuntimeListResponse
	if err := json.Unmarshal(afterMoveDevicesResp.Body.Bytes(), &afterMoveDevices); err != nil {
		t.Fatalf("after move devices json: %v", err)
	}
	originalAfterMove, ok := findDevice(afterMoveDevices.Devices, f.enrollment.DeviceID)
	if !ok {
		t.Fatalf("original device missing after move: %+v", afterMoveDevices.Devices)
	}
	if _, ok := findRuntime(originalAfterMove.Runtimes, f.codexRuntimeID); ok {
		t.Fatalf("moved runtime must be removed from original device: %+v", originalAfterMove.Runtimes)
	}
	replacementAfterMove, ok := findDevice(afterMoveDevices.Devices, replacement.DeviceID)
	if !ok {
		t.Fatalf("replacement device missing after move: %+v", afterMoveDevices.Devices)
	}
	if _, ok := findRuntime(replacementAfterMove.Runtimes, f.codexRuntimeID); !ok {
		t.Fatalf("moved runtime missing from replacement device: %+v", replacementAfterMove.Runtimes)
	}
}

func (f *deviceEnrollmentHTTPFixture) verifyReplacementDeviceBindings(replacement EnrollDeviceResponse) {
	t := f.t
	replacementBindingsReq := httptest.NewRequest(http.MethodGet, "/v1/daemon/agent-bindings", nil)
	replacementBindingsReq.Header.Set(deviceIDHeader, replacement.DeviceID)
	replacementBindingsReq.Header.Set(deviceSecretHeader, replacement.DeviceSecret)
	replacementBindingsResp := httptest.NewRecorder()
	f.server.ServeHTTP(replacementBindingsResp, replacementBindingsReq)
	if replacementBindingsResp.Code != http.StatusOK {
		t.Fatalf("replacement bindings status=%d body=%s", replacementBindingsResp.Code, replacementBindingsResp.Body.String())
	}
	var replacementBindings AgentRuntimeBindingListResponse
	if err := json.Unmarshal(replacementBindingsResp.Body.Bytes(), &replacementBindings); err != nil {
		t.Fatalf("replacement bindings json: %v", err)
	}
	if len(replacementBindings.Bindings) != 1 ||
		replacementBindings.Bindings[0].AgentID != f.created.Agent.AgentID ||
		replacementBindings.Bindings[0].DeviceID != replacement.DeviceID ||
		replacementBindings.Bindings[0].RuntimeID != f.codexRuntimeID {
		t.Fatalf("replacement bindings = %+v", replacementBindings.Bindings)
	}
}
