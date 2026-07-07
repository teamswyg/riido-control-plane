package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPDaemonAgentBindingsSuccess(t *testing.T) {
	runtimeStore := &daemonBindingsRuntimeStore{
		response: AgentRuntimeBindingListResponse{
			SchemaVersion: SchemaVersion,
			Bindings: []AgentRuntimeBinding{{
				AgentID:         "agent-a",
				DeviceID:        "device-a",
				RuntimeID:       "runtime-a",
				RuntimeProvider: "codex",
			}},
		},
	}
	resp := serveDaemonBindings(newDaemonBindingsTestServer(runtimeStore), "device-a", "secret")
	if resp.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d body=%s", resp.Code, http.StatusOK, resp.Body.String())
	}
	if runtimeStore.gotID != "device-a" {
		t.Fatalf("runtime store device id=%q", runtimeStore.gotID)
	}
	var out AgentRuntimeBindingListResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode bindings: %v", err)
	}
	if len(out.Bindings) != 1 || out.Bindings[0].RuntimeID != "runtime-a" {
		t.Fatalf("bindings = %+v", out.Bindings)
	}
}
