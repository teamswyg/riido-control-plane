package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (f *deviceEnrollmentHTTPFixture) syncCursorRuntimeAndVerifyMergedReadModels() {
	t := f.t
	cursorSnapshotReq := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(`{"daemon_id":"daemon-enrolled","runtimes":[{"runtime_id":"`+f.cursorRuntimeID+`","kind":"cursor"}]}`))
	cursorSnapshotReq.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
	cursorSnapshotReq.Header.Set(deviceSecretHeader, f.enrollment.DeviceSecret)
	cursorSnapshotResp := httptest.NewRecorder()
	f.server.ServeHTTP(cursorSnapshotResp, cursorSnapshotReq)
	if cursorSnapshotResp.Code != http.StatusAccepted {
		t.Fatalf("cursor snapshot status=%d body=%s", cursorSnapshotResp.Code, cursorSnapshotResp.Body.String())
	}
	mergedDevicesReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices", nil)
	mergedDevicesReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	mergedDevicesResp := httptest.NewRecorder()
	f.server.ServeHTTP(mergedDevicesResp, mergedDevicesReq)
	if mergedDevicesResp.Code != http.StatusOK {
		t.Fatalf("merged devices status=%d body=%s", mergedDevicesResp.Code, mergedDevicesResp.Body.String())
	}
	var mergedDevices DeviceRuntimeListResponse
	if err := json.Unmarshal(mergedDevicesResp.Body.Bytes(), &mergedDevices); err != nil {
		t.Fatalf("merged devices json: %v", err)
	}
	mergedDevice, ok := findDevice(mergedDevices.Devices, f.enrollment.DeviceID)
	if !ok {
		t.Fatalf("merged enrolled device missing: %+v", mergedDevices.Devices)
	}
	codexRuntime, ok := findRuntime(mergedDevice.Runtimes, f.codexRuntimeID)
	if !ok || !codexRuntime.RequiresExperimentalOptIn {
		t.Fatalf("codex runtime opt-in fact missing after second snapshot: %+v", mergedDevice.Runtimes)
	}
	if _, ok := findRuntime(mergedDevice.Runtimes, f.cursorRuntimeID); !ok {
		t.Fatalf("cursor runtime missing after second snapshot: %+v", mergedDevice.Runtimes)
	}
}
