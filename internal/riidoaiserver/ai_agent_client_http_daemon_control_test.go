package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentDeviceDaemonDetailAndControl(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{
		{
			PrincipalID: "user-1",
			Token:       "user-token",
			Scopes:      []string{"ai-agent:*"},
		},
		{
			PrincipalID: "admin-user",
			Token:       "admin-token",
			Scopes:      []string{"ai-agent:*"},
			Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
		},
	})

	detailReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-owned-codex/daemon", nil)
	detailReq.Header.Set("Authorization", "Bearer user-token")
	detailResp := httptest.NewRecorder()
	server.ServeHTTP(detailResp, detailReq)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("daemon detail status=%d body=%s", detailResp.Code, detailResp.Body.String())
	}
	var detail DeviceDaemonDetailResponse
	if err := json.Unmarshal(detailResp.Body.Bytes(), &detail); err != nil {
		t.Fatalf("daemon detail json: %v", err)
	}
	if detail.Daemon.Availability != DaemonAvailabilityOnline || detail.Daemon.PID != 5111 || detail.Daemon.Profile != "desktop-api.riido.ai" {
		t.Fatalf("daemon detail = %+v", detail.Daemon)
	}
	if detail.Runtime == nil ||
		detail.Runtime.RuntimeID != "runtime-codex-dev" ||
		detail.Runtime.ProviderVersion != "codex-cli 0.133.0" {
		t.Fatalf("daemon detail runtime = %+v", detail.Runtime)
	}
	if !sameDaemonActions(detail.Daemon.SupportedActions, []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop}) {
		t.Fatalf("daemon supported actions = %+v", detail.Daemon.SupportedActions)
	}

	publicDetailReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon", nil)
	publicDetailReq.Header.Set("Authorization", "Bearer user-token")
	publicDetailResp := httptest.NewRecorder()
	server.ServeHTTP(publicDetailResp, publicDetailReq)
	if publicDetailResp.Code != http.StatusOK {
		t.Fatalf("public agent daemon detail status=%d body=%s", publicDetailResp.Code, publicDetailResp.Body.String())
	}
	var publicDetail DeviceDaemonDetailResponse
	if err := json.Unmarshal(publicDetailResp.Body.Bytes(), &publicDetail); err != nil {
		t.Fatalf("public daemon detail json: %v", err)
	}
	if publicDetail.Daemon.DeviceID != "device-shared-studio" || publicDetail.Daemon.OwnerPrincipalID != "user-2" {
		t.Fatalf("public agent daemon detail = %+v", publicDetail.Daemon)
	}
	if publicDetail.Runtime == nil ||
		publicDetail.Runtime.RuntimeID != "runtime-openclaw-shared" ||
		publicDetail.Runtime.ProviderVersion != "openclaw 0.1.0" {
		t.Fatalf("public daemon detail runtime = %+v", publicDetail.Runtime)
	}

	privateDeniedReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-private-cursor/daemon", nil)
	privateDeniedReq.Header.Set("Authorization", "Bearer user-token")
	privateDeniedResp := httptest.NewRecorder()
	server.ServeHTTP(privateDeniedResp, privateDeniedReq)
	if privateDeniedResp.Code != http.StatusNotFound {
		t.Fatalf("private agent daemon detail for non-admin status=%d body=%s", privateDeniedResp.Code, privateDeniedResp.Body.String())
	}

	adminPrivateReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-private-cursor/daemon", nil)
	adminPrivateReq.Header.Set("Authorization", "Bearer admin-token")
	adminPrivateResp := httptest.NewRecorder()
	server.ServeHTTP(adminPrivateResp, adminPrivateReq)
	if adminPrivateResp.Code != http.StatusOK {
		t.Fatalf("private agent daemon detail for admin status=%d body=%s", adminPrivateResp.Code, adminPrivateResp.Body.String())
	}

	restartReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon/restart", strings.NewReader(`{"reason":"settings page restart"}`))
	restartReq.Header.Set("Authorization", "Bearer user-token")
	restartResp := httptest.NewRecorder()
	server.ServeHTTP(restartResp, restartReq)
	if restartResp.Code != http.StatusAccepted {
		t.Fatalf("daemon restart status=%d body=%s", restartResp.Code, restartResp.Body.String())
	}
	var restart DeviceDaemonCommandResponse
	if err := json.Unmarshal(restartResp.Body.Bytes(), &restart); err != nil {
		t.Fatalf("daemon restart json: %v", err)
	}
	if restart.Action != DaemonControlActionRestart || restart.ControlState != DaemonControlStateRestarting {
		t.Fatalf("daemon restart = %+v", restart)
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon/stop", strings.NewReader(`{"reason":"settings page stop"}`))
	stopReq.Header.Set("Authorization", "Bearer user-token")
	stopResp := httptest.NewRecorder()
	server.ServeHTTP(stopResp, stopReq)
	if stopResp.Code != http.StatusAccepted {
		t.Fatalf("daemon stop status=%d body=%s", stopResp.Code, stopResp.Body.String())
	}
	var stop DeviceDaemonCommandResponse
	if err := json.Unmarshal(stopResp.Body.Bytes(), &stop); err != nil {
		t.Fatalf("daemon stop json: %v", err)
	}
	if stop.Action != DaemonControlActionStop || stop.Availability != DaemonAvailabilityOffline || stop.ControlState != DaemonControlStateStopping {
		t.Fatalf("daemon stop = %+v", stop)
	}

	devicesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/devices", nil)
	devicesReq.Header.Set("Authorization", "Bearer user-token")
	devicesResp := httptest.NewRecorder()
	server.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("devices after daemon stop status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("devices json: %v", err)
	}
	var sharedDevice DeviceRecord
	for _, device := range devices.Devices {
		if device.DeviceID == "device-shared-studio" {
			sharedDevice = device
			break
		}
	}
	if sharedDevice.DeviceID == "" || len(sharedDevice.Runtimes) != 1 {
		t.Fatalf("shared public-agent device should be visible with one public runtime: %+v", devices.Devices)
	}
	for _, runtime := range sharedDevice.Runtimes {
		if runtime.RuntimeID != "runtime-openclaw-shared" || runtime.Availability != RuntimeAvailabilityOffline {
			t.Fatalf("public runtime should be offline after daemon stop: %+v", runtime)
		}
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	server.ServeHTTP(eventsResp, eventsReq)
	if eventsResp.Code != http.StatusOK || !strings.Contains(eventsResp.Body.String(), AgentClientEventDeviceDaemonStatus) {
		t.Fatalf("events should include daemon status, status=%d body=%s", eventsResp.Code, eventsResp.Body.String())
	}

	privateStopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents/agent-private-cursor/daemon/stop", nil)
	privateStopReq.Header.Set("Authorization", "Bearer user-token")
	privateStopResp := httptest.NewRecorder()
	server.ServeHTTP(privateStopResp, privateStopReq)
	if privateStopResp.Code != http.StatusNotFound {
		t.Fatalf("private agent daemon control for non-admin status=%d body=%s", privateStopResp.Code, privateStopResp.Body.String())
	}
}
