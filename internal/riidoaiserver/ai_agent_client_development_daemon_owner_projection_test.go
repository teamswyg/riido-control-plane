package riidoaiserver

import "testing"

func TestDeviceDaemonForOwnerAllowsDeviceOwnerWhenDaemonOwnerDrifts(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{PrincipalID: "user-1"}

	store.mu.Lock()
	for key, daemon := range store.daemons {
		if daemon.DeviceID == "device-dev-macbook" {
			daemon.OwnerPrincipalID = "stale-daemon-owner"
			store.daemons[key] = daemon
		}
	}
	daemon, ok := store.deviceDaemonForOwnerLocked(owner, "device-dev-macbook")
	store.mu.Unlock()

	if !ok {
		t.Fatal("deviceDaemonForOwnerLocked() did not allow owning device read model")
	}
	if daemon.OwnerPrincipalID != "stale-daemon-owner" {
		t.Fatalf("daemon owner = %q", daemon.OwnerPrincipalID)
	}
}

func TestDeviceDaemonForOwnerRejectsOwnedAgentOnDifferentDevice(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "agent-runtime-owner-wrong-device"}

	store.mu.Lock()
	agent := store.agents["agent-owned-codex"]
	agent.OwnerPrincipalID = principal.PrincipalID
	agent.RuntimeID = "runtime-codex-dev"
	store.agents[agent.AgentID] = agent
	_, ok := store.deviceDaemonForOwnerLocked(principal, "device-shared-runner")
	store.mu.Unlock()

	if ok {
		t.Fatal("deviceDaemonForOwnerLocked() allowed owned runtime on different device")
	}
}
