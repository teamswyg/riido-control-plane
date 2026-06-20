package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"
)

func (f *deviceEnrollmentHTTPFixture) projectStaleRuntimeAndDaemonOffline() {
	f.markEnrolledDeviceStale()
	f.verifyStaleDeviceProjection()
	f.verifyStaleDaemonProjection()
	f.verifyStaleRuntimeExcludedFromBindings()
}

func (f *deviceEnrollmentHTTPFixture) markEnrolledDeviceStale() {
	staleAt := time.Now().UTC().Add(-AIAgentDeviceRuntimeSnapshotStaleAfter - time.Second)
	f.aiAgentStore.mu.Lock()
	for deviceIndex := range f.aiAgentStore.devices {
		if f.aiAgentStore.devices[deviceIndex].DeviceID == f.enrollment.DeviceID {
			f.aiAgentStore.devices[deviceIndex].DaemonLastSeenAt = staleAt
			break
		}
	}
	if daemon, ok := f.aiAgentStore.daemons[f.enrollment.DeviceID]; ok {
		daemon.LastSeenAt = staleAt
		f.aiAgentStore.daemons[f.enrollment.DeviceID] = daemon
	}
	f.aiAgentStore.mu.Unlock()
}

func (f *deviceEnrollmentHTTPFixture) verifyStaleDeviceProjection() {
	t := f.t
	staleDevicesReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices", nil)
	staleDevicesReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	staleDevicesResp := httptest.NewRecorder()
	f.server.ServeHTTP(staleDevicesResp, staleDevicesReq)
	if staleDevicesResp.Code != http.StatusOK {
		t.Fatalf("stale devices status=%d body=%s", staleDevicesResp.Code, staleDevicesResp.Body.String())
	}
	var staleDevices DeviceRuntimeListResponse
	if err := json.Unmarshal(staleDevicesResp.Body.Bytes(), &staleDevices); err != nil {
		t.Fatalf("stale devices json: %v", err)
	}
	staleDevice, ok := findDevice(staleDevices.Devices, f.enrollment.DeviceID)
	if !ok {
		t.Fatalf("stale device missing: %+v", staleDevices.Devices)
	}
	staleCodex, ok := findRuntime(staleDevice.Runtimes, f.codexRuntimeID)
	if !ok || staleCodex.Availability != RuntimeAvailabilityOffline || staleCodex.DetectionState != RuntimeDetectionStateMissing {
		t.Fatalf("stale codex runtime should project offline/missing: %+v", staleDevice.Runtimes)
	}
}
