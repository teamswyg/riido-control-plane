package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAgentBindingRemainsAvailableAcrossLegacyHeartbeatJitter(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: "workspace-a"}
	enrollment := enrollTestDevice(t, ctx, store, principal, principal.WorkspaceID, "machine-legacy-heartbeat", "Legacy Mac")
	runtimeID := enrollment.DeviceID + ":cursor"
	syncTestRuntime(t, ctx, store, principal, enrollment.DeviceID, enrollment.DeviceID, runtimeID)
	created, err := store.createAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "Legacy Cursor",
		RuntimeID:  runtimeID,
		Visibility: AgentVisibilityPrivate,
	}, "")
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	// Older daemons can send the next snapshot just after a 20-second poll.
	// That timing is delayed, not disconnected, and must keep the binding alive.
	lastSeenAt := time.Now().UTC().Add(-21 * time.Second)
	store.mu.Lock()
	for i := range store.devices {
		if store.devices[i].DeviceID == enrollment.DeviceID {
			store.devices[i].DaemonLastSeenAt = lastSeenAt
		}
	}
	for key, daemon := range store.daemons {
		if daemon.DeviceID == enrollment.DeviceID {
			daemon.LastSeenAt = lastSeenAt
			store.daemons[key] = daemon
		}
	}
	store.mu.Unlock()

	binding, ok := store.LookupAgent(created.Agent.AgentID)
	if !ok || binding.RuntimeID != runtimeID || binding.DeviceID != enrollment.DeviceID {
		t.Fatalf("binding lost across legacy heartbeat jitter: binding=%+v found=%v", binding, ok)
	}
}
