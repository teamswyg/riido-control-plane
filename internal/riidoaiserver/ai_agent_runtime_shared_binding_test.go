package riidoaiserver

import (
	"context"
	"testing"
)

func TestDaemonAgentBindingsIncludeAgentsSharingRuntime(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: "workspace-a"}
	enroll, err := store.EnrollDeviceCredential(ctx, principal, "workspace-a",
		EnrollDeviceRequest{MachineID: "machine-shared-runtime", DisplayName: "Shared Runtime Mac"})
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	runtimeID := "machine-shared-runtime:claude"
	_, err = store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID: "machine-shared-runtime",
		DeviceID: enroll.DeviceID,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:      runtimeID,
			Kind:           RuntimeKindClaudeCode,
			Availability:   RuntimeAvailabilityOnline,
			DetectionState: RuntimeDetectionStateDetected,
		}},
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	first := createAgentForSharedRuntime(t, store, ctx, principal, runtimeID, "First")
	second := createAgentForSharedRuntime(t, store, ctx, principal, runtimeID, "Second")
	bindings, err := store.ListAIAgentDaemonAgentBindings(ctx,
		AuthorizationResult{PrincipalID: "owner-user"}, enroll.DeviceID)
	if err != nil {
		t.Fatalf("ListAIAgentDaemonAgentBindings: %v", err)
	}
	seen := map[string]bool{}
	for _, binding := range bindings.Bindings {
		if binding.RuntimeID == runtimeID {
			seen[binding.AgentID] = true
		}
	}
	if !seen[first.Agent.AgentID] || !seen[second.Agent.AgentID] {
		t.Fatalf("shared runtime bindings = %+v", bindings.Bindings)
	}
}

func createAgentForSharedRuntime(t *testing.T, store *DevelopmentAIAgentClientStore, ctx context.Context, principal AuthorizationResult, runtimeID, name string) AgentClientRecordResponse {
	t.Helper()
	created, err := store.createAIAgent(ctx, principal,
		CreateAgentConfigurationRequest{Name: name, RuntimeID: runtimeID, Visibility: AgentVisibilityPrivate}, "")
	if err != nil {
		t.Fatalf("create agent %s: %v", name, err)
	}
	return created
}
