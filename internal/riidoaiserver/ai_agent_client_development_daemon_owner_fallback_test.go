package riidoaiserver

import "testing"

func TestDeviceDaemonForOwnerFallsBackThroughAgentRuntime(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "agent-runtime-owner-fallback"}

	store.mu.Lock()
	agent := store.agents["agent-owned-codex"]
	agent.OwnerPrincipalID = principal.PrincipalID
	agent.RuntimeID = "runtime-codex-dev"
	store.agents[agent.AgentID] = agent
	store.daemons = map[string]DeviceDaemonRecord{}

	daemon, ok := store.deviceDaemonForOwnerLocked(principal, "device-dev-macbook")
	store.mu.Unlock()

	if !ok {
		t.Fatal("deviceDaemonForOwnerLocked() did not fall back through owned agent runtime")
	}
	if daemon.DeviceID != "device-dev-macbook" {
		t.Fatalf("fallback daemon device_id = %q", daemon.DeviceID)
	}
}
