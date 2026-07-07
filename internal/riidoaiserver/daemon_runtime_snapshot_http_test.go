package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPDaemonRuntimeSnapshotSyncsDeviceCredentialSnapshot(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	if err := store.ConfigureDaemonProfile("staging"); err != nil {
		t.Fatalf("ConfigureDaemonProfile: %v", err)
	}
	server := newRuntimeSnapshotHTTPTestServer(t, store)
	enrollment := enrollRuntimeSnapshotDevice(t, server)

	body := `{"device_id":"other","daemon_id":"daemon-a","runtimes":[]}`
	req := runtimeSnapshotRequest(body, enrollment)
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("mismatch status=%d body=%s", resp.Code, resp.Body.String())
	}

	req = runtimeSnapshotRequest(`{`, enrollment)
	resp = httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("bad json status=%d body=%s", resp.Code, resp.Body.String())
	}

	body = `{"daemon_id":"daemon-a","profile":"staging","app_version":"v1.26.2","runtimes":[{"runtime_id":"runtime-codex","kind":"codex","availability":"online","detection_state":"detected"}]}`
	req = runtimeSnapshotRequest(body, enrollment)
	resp = httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("sync status=%d body=%s", resp.Code, resp.Body.String())
	}
	var synced DeviceRuntimeSnapshotSyncResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &synced); err != nil {
		t.Fatalf("sync json: %v", err)
	}
	if synced.Device.DeviceID != enrollment.DeviceID || synced.Daemon.AppVersion != "v1.26.2" {
		t.Fatalf("synced snapshot = %+v", synced)
	}
}
