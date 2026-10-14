package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func (f *deviceEnrollmentHTTPFixture) verifyStaleDaemonProjection() {
	t := f.t
	staleDaemonReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/agents/"+f.created.Agent.AgentID+"/daemon", nil)
	staleDaemonReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	staleDaemonResp := httptest.NewRecorder()
	f.server.ServeHTTP(staleDaemonResp, staleDaemonReq)
	if staleDaemonResp.Code != http.StatusOK {
		t.Fatalf("stale daemon detail status=%d body=%s", staleDaemonResp.Code, staleDaemonResp.Body.String())
	}
	var staleDaemon DeviceDaemonDetailResponse
	if err := json.Unmarshal(staleDaemonResp.Body.Bytes(), &staleDaemon); err != nil {
		t.Fatalf("stale daemon json: %v", err)
	}
	if staleDaemon.Daemon.Availability != DaemonAvailabilityOffline ||
		staleDaemon.Daemon.ControlState != DaemonControlStateIdle ||
		!daemonSupportsAction(staleDaemon.Daemon, DaemonControlActionStart) ||
		daemonSupportsAction(staleDaemon.Daemon, DaemonControlActionStop) {
		t.Fatalf("stale daemon should project offline/start-only: %+v", staleDaemon.Daemon)
	}
}

func (f *deviceEnrollmentHTTPFixture) verifyStaleRuntimeExcludedFromBindings() {
	t := f.t
	staleBindingsReq := httptest.NewRequest(http.MethodGet, "/v1/daemon/agent-bindings", nil)
	staleBindingsReq.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
	staleBindingsReq.Header.Set(deviceSecretHeader, f.enrollment.DeviceSecret)
	staleBindingsResp := httptest.NewRecorder()
	f.server.ServeHTTP(staleBindingsResp, staleBindingsReq)
	if staleBindingsResp.Code != http.StatusOK {
		t.Fatalf("stale bindings status=%d body=%s", staleBindingsResp.Code, staleBindingsResp.Body.String())
	}
	var staleBindings AgentRuntimeBindingListResponse
	if err := json.Unmarshal(staleBindingsResp.Body.Bytes(), &staleBindings); err != nil {
		t.Fatalf("stale bindings json: %v", err)
	}
	if len(staleBindings.Bindings) != 0 {
		t.Fatalf("stale runtime must be excluded from daemon bindings: %+v", staleBindings.Bindings)
	}
}
