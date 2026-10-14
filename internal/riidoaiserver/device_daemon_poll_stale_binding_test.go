package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDaemonPollUsesDurableBindingAcrossStaleSnapshot(t *testing.T) {
	ctx := context.Background()
	registry := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: "workspace-a"}
	enrollment := enrollTestDevice(t, ctx, registry, principal, principal.WorkspaceID, "machine-stale-poll", "Stale Poll Mac")
	runtimeID := enrollment.DeviceID + ":cursor"
	syncTestRuntime(t, ctx, registry, principal, enrollment.DeviceID, enrollment.DeviceID, runtimeID)
	created, err := registry.createAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name: "Stale Poll Cursor", RuntimeID: runtimeID, Visibility: AgentVisibilityPrivate,
	}, "")
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	lastSeenAt := time.Now().UTC().Add(-AIAgentDeviceRuntimeSnapshotStaleAfter - time.Second)
	registry.mu.Lock()
	for i := range registry.devices {
		if registry.devices[i].DeviceID == enrollment.DeviceID {
			registry.devices[i].DaemonLastSeenAt = lastSeenAt
		}
	}
	for key, daemon := range registry.daemons {
		if daemon.DeviceID == enrollment.DeviceID {
			daemon.LastSeenAt = lastSeenAt
			registry.daemons[key] = daemon
		}
	}
	registry.mu.Unlock()

	if _, ok := registry.LookupAgent(created.Agent.AgentID); ok {
		t.Fatal("stale snapshot must remain unavailable to liveness-sensitive lookup")
	}
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: registry})
	defer assignmentStore.Close()
	poll, err := assignmentStore.PollAgent(ctx, created.Agent.AgentID, PollRequest{
		DaemonID: enrollment.DeviceID, DeviceID: enrollment.DeviceID, RuntimeID: runtimeID,
	})
	if err != nil || poll.Action != PollNone {
		t.Fatalf("durable daemon poll = %+v err=%v, want PollNone", poll, err)
	}
	if _, err := assignmentStore.PollAgent(ctx, created.Agent.AgentID, PollRequest{
		DaemonID: "wrong-daemon", DeviceID: enrollment.DeviceID, RuntimeID: runtimeID,
	}); !errors.Is(err, ErrAgentBindingValidation) {
		t.Fatalf("wrong daemon poll err=%v, want ErrAgentBindingValidation", err)
	}
}
