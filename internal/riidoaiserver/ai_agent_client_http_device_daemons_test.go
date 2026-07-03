package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPAIAgentClientDeviceDaemonsListsAllDaemonProfiles(t *testing.T) {
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	now := time.Date(2026, 7, 3, 9, 0, 0, 0, time.UTC)
	aiAgentStore.mu.Lock()
	aiAgentStore.putDaemonLocked(DeviceDaemonRecord{
		DeviceID:          "device-dev-macbook",
		OwnerPrincipalID:  "user-1",
		DeviceDisplayName: "JY MacBook",
		DaemonID:          "daemon-dev-macbook-staging",
		Profile:           "staging",
		PID:               7001,
		StartedAt:         now.Add(-30 * time.Minute),
		LastSeenAt:        now,
		Availability:      DaemonAvailabilityOnline,
		ControlState:      DaemonControlStateIdle,
	})
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
	if list.DeviceID != "device-dev-macbook" || len(list.Daemons) != 2 {
		t.Fatalf("device daemons = %+v", list)
	}
	if list.Daemons[0].Profile != "staging" || list.Daemons[0].PID != 7001 {
		t.Fatalf("latest daemon should sort first: %+v", list.Daemons)
	}
}

func newAIAgentClientHTTPTestServerWithStore(t *testing.T, aiAgentStore *DevelopmentAIAgentClientStore) http.Handler {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	return NewServer(ServerConfig{AIAgentClient: aiAgentStore, Assignment: assignmentStore, Authorizer: authorizer}).Handler()
}
