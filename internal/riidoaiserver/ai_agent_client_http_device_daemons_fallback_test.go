package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientDeviceDaemonsIncludesReadModelFallback(t *testing.T) {
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	aiAgentStore.mu.Lock()
	aiAgentStore.daemons = map[string]DeviceDaemonRecord{}
	aiAgentStore.mu.Unlock()
	server := newAIAgentClientHTTPTestServerWithStore(t, aiAgentStore)

	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/daemons", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("device daemons status=%d body=%s", resp.Code, resp.Body.String())
	}
	var list DeviceDaemonListResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &list); err != nil {
		t.Fatalf("device daemons json: %v", err)
	}
	if len(list.Daemons) != 1 {
		t.Fatalf("device daemons fallback count = %d, want 1", len(list.Daemons))
	}
	daemon := list.Daemons[0]
	if daemon.DeviceID != "device-dev-macbook" {
		t.Fatalf("fallback daemon device_id = %q", daemon.DeviceID)
	}
	if len(daemon.SupportedActions) == 0 {
		t.Fatalf("fallback daemon supported_actions is empty: %+v", daemon)
	}
}
