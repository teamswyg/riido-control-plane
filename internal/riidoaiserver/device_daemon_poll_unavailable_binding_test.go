package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestDaemonPollRejectsExplicitlyUnavailableRuntime(t *testing.T) {
	ctx := context.Background()
	registry := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: "workspace-a"}
	enrollment := enrollTestDevice(t, ctx, registry, principal, principal.WorkspaceID, "machine-offline-poll", "Offline Poll Mac")
	runtimeID := enrollment.DeviceID + ":cursor"
	syncTestRuntime(t, ctx, registry, principal, enrollment.DeviceID, enrollment.DeviceID, runtimeID)
	created, err := registry.createAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name: "Offline Poll Cursor", RuntimeID: runtimeID, Visibility: AgentVisibilityPrivate,
	}, "")
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	registry.mu.Lock()
	for i := range registry.devices {
		for j := range registry.devices[i].Runtimes {
			if registry.devices[i].Runtimes[j].RuntimeID == runtimeID {
				registry.devices[i].Runtimes[j].Availability = RuntimeAvailabilityOffline
				registry.devices[i].Runtimes[j].DetectionState = RuntimeDetectionStateMissing
			}
		}
	}
	registry.mu.Unlock()

	err = validateDaemonBinding(registry, created.Agent.AgentID, PollRequest{
		DaemonID: enrollment.DeviceID, DeviceID: enrollment.DeviceID, RuntimeID: runtimeID,
	})
	if !errors.Is(err, ErrAgentBindingValidation) {
		t.Fatalf("explicitly unavailable runtime err=%v, want ErrAgentBindingValidation", err)
	}
}
