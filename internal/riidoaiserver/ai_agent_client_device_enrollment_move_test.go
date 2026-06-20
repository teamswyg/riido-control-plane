package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (f *deviceEnrollmentHTTPFixture) moveRuntimeToReplacementDevice() {
	replacement := f.enrollReplacementDevice()
	f.syncMovedCodexRuntime(replacement)
	f.verifyRuntimeMovedToReplacementDevice(replacement)
	f.verifyReplacementDeviceBindings(replacement)
}

func (f *deviceEnrollmentHTTPFixture) enrollReplacementDevice() EnrollDeviceResponse {
	t := f.t
	replacementReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", strings.NewReader(`{"display_name":"Replacement Mac","platform":"darwin","app_version":"0.0.2"}`))
	replacementReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	replacementResp := httptest.NewRecorder()
	f.server.ServeHTTP(replacementResp, replacementReq)
	if replacementResp.Code != http.StatusCreated {
		t.Fatalf("replacement enroll status=%d body=%s", replacementResp.Code, replacementResp.Body.String())
	}
	var replacement EnrollDeviceResponse
	if err := json.Unmarshal(replacementResp.Body.Bytes(), &replacement); err != nil {
		t.Fatalf("replacement json: %v", err)
	}
	if replacement.DeviceID == "" || replacement.DeviceID == f.enrollment.DeviceID {
		t.Fatalf("replacement enrollment = %+v", replacement)
	}
	return replacement
}

func (f *deviceEnrollmentHTTPFixture) syncMovedCodexRuntime(replacement EnrollDeviceResponse) {
	t := f.t
	moveSnapshotReq := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(`{"daemon_id":"daemon-replacement","runtimes":[{"runtime_id":"`+f.codexRuntimeID+`","kind":"codex","requires_experimental_opt_in":true}]}`))
	moveSnapshotReq.Header.Set(deviceIDHeader, replacement.DeviceID)
	moveSnapshotReq.Header.Set(deviceSecretHeader, replacement.DeviceSecret)
	moveSnapshotResp := httptest.NewRecorder()
	f.server.ServeHTTP(moveSnapshotResp, moveSnapshotReq)
	if moveSnapshotResp.Code != http.StatusAccepted {
		t.Fatalf("move snapshot status=%d body=%s", moveSnapshotResp.Code, moveSnapshotResp.Body.String())
	}
}
