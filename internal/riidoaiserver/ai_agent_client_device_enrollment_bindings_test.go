package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func (f *deviceEnrollmentHTTPFixture) verifyDaemonBindingsAfterRuntimeSnapshot() {
	t := f.t
	bindingsReq := httptest.NewRequest(http.MethodGet, "/v1/daemon/agent-bindings", nil)
	bindingsReq.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
	bindingsReq.Header.Set(deviceSecretHeader, f.enrollment.DeviceSecret)
	bindingsResp := httptest.NewRecorder()
	f.server.ServeHTTP(bindingsResp, bindingsReq)
	if bindingsResp.Code != http.StatusOK {
		t.Fatalf("bindings status=%d body=%s", bindingsResp.Code, bindingsResp.Body.String())
	}
	var bindings AgentRuntimeBindingListResponse
	if err := json.Unmarshal(bindingsResp.Body.Bytes(), &bindings); err != nil {
		t.Fatalf("bindings json: %v", err)
	}
	if len(bindings.Bindings) != 1 ||
		bindings.Bindings[0].AgentID != f.created.Agent.AgentID ||
		bindings.Bindings[0].RuntimeID != f.codexRuntimeID ||
		bindings.Bindings[0].RuntimeProvider != "codex" {
		t.Fatalf("bindings after second runtime snapshot = %+v", bindings.Bindings)
	}
}
