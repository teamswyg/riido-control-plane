package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (f *deviceEnrollmentHTTPFixture) syncCodexRuntimeAndCreateAgent() {
	t := f.t
	f.codexRuntimeID = f.enrollment.DeviceID + ":codex"
	f.cursorRuntimeID = f.enrollment.DeviceID + ":cursor"
	codexSnapshotReq := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(`{"daemon_id":"daemon-enrolled","runtimes":[{"runtime_id":"`+f.codexRuntimeID+`","kind":"codex","requires_experimental_opt_in":true}]}`))
	codexSnapshotReq.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
	codexSnapshotReq.Header.Set(deviceSecretHeader, f.enrollment.DeviceSecret)
	codexSnapshotResp := httptest.NewRecorder()
	f.server.ServeHTTP(codexSnapshotResp, codexSnapshotReq)
	if codexSnapshotResp.Code != http.StatusAccepted {
		t.Fatalf("codex snapshot status=%d body=%s", codexSnapshotResp.Code, codexSnapshotResp.Body.String())
	}
	createBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:       "enrolled codex",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  f.codexRuntimeID,
	})
	if err != nil {
		t.Fatalf("marshal create body: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-alpha/ai-agent/agents", strings.NewReader(string(createBody)))
	createReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	createResp := httptest.NewRecorder()
	f.server.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	if err := json.Unmarshal(createResp.Body.Bytes(), &f.created); err != nil {
		t.Fatalf("create json: %v", err)
	}
	if f.created.Agent.RuntimeID != f.codexRuntimeID || f.created.Agent.RuntimeKind != RuntimeKindCodex {
		t.Fatalf("created agent = %+v", f.created.Agent)
	}
}
