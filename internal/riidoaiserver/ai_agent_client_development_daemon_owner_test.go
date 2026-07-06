package riidoaiserver

import "testing"

func TestDeviceDaemonForOwnerAllowsAgentRuntimeOwner(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "agent-runtime-owner"}

	store.mu.Lock()
	agent := store.agents["agent-owned-codex"]
	agent.OwnerPrincipalID = principal.PrincipalID
	agent.RuntimeID = "runtime-codex-dev"
	store.agents[agent.AgentID] = agent

	daemon, ok := store.deviceDaemonForOwnerLocked(principal, "device-dev-macbook")
	store.mu.Unlock()

	if !ok {
		t.Fatal("deviceDaemonForOwnerLocked() did not find daemon through owned agent runtime")
	}
	if daemon.DeviceID != "device-dev-macbook" {
		t.Fatalf("daemon device_id = %q", daemon.DeviceID)
	}
}

func TestDeviceByRuntimeIDLockedReturnsCopiedDevice(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()

	store.mu.Lock()
	device, ok := store.deviceByRuntimeIDLocked("runtime-codex-dev")
	device.Runtimes[0].RuntimeID = "mutated"
	again, okAgain := store.deviceByRuntimeIDLocked("runtime-codex-dev")
	_, missing := store.deviceByRuntimeIDLocked("runtime-missing")
	store.mu.Unlock()

	if !ok || !okAgain {
		t.Fatal("deviceByRuntimeIDLocked() did not find seed runtime")
	}
	if device.DeviceID != "device-dev-macbook" {
		t.Fatalf("device_id = %q", device.DeviceID)
	}
	if again.Runtimes[0].RuntimeID == "mutated" {
		t.Fatal("deviceByRuntimeIDLocked() returned mutable internal runtime slice")
	}
	if missing {
		t.Fatal("deviceByRuntimeIDLocked() found unknown runtime")
	}
}

func TestDeviceDaemonForOwnerCoversOwnerFallbackAndDenied(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{PrincipalID: "user-1"}

	store.mu.Lock()
	daemon, ok := store.deviceDaemonForOwnerLocked(owner, "device-dev-macbook")
	store.daemons = map[string]DeviceDaemonRecord{}
	fallback, fallbackOK := store.deviceDaemonForOwnerLocked(owner, "device-dev-macbook")
	_, denied := store.deviceDaemonForOwnerLocked(
		AuthorizationResult{PrincipalID: "not-device-or-agent-owner"},
		"device-dev-macbook",
	)
	store.mu.Unlock()

	if !ok || daemon.DeviceID != "device-dev-macbook" {
		t.Fatalf("owned daemon lookup = %+v, %v", daemon, ok)
	}
	if !fallbackOK || fallback.DeviceID != "device-dev-macbook" {
		t.Fatalf("device fallback daemon lookup = %+v, %v", fallback, fallbackOK)
	}
	if denied {
		t.Fatal("deviceDaemonForOwnerLocked() allowed unrelated principal")
	}
}
