package riidoaiserver

import (
	"context"
	"testing"
)

func TestPersistentAIAgentClientStoreMutationDelegatesSave(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	device, err := store.ConnectAIAgentDevice(ctx, principal, "machine-persist-delegate")
	if err != nil || device.DeviceID == "" {
		t.Fatalf("ConnectAIAgentDevice = %+v, %v", device, err)
	}
	afterConnect := snapshots.saves
	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:        "delegate persistence agent",
		Description: stringPtr("delegate persistence description"),
		Instruction: stringPtr("delegate persistence instruction"),
		Visibility:  AgentVisibilityPrivate,
		RuntimeID:   "runtime-cursor-dev",
		ModelID:     stringPtr("cursor-auto"),
	})
	if err != nil || created.Agent.AgentID == "" {
		t.Fatalf("CreateAIAgent = %+v, %v", created, err)
	}
	if snapshots.saves <= afterConnect {
		t.Fatalf("CreateAIAgent did not save snapshot: %d <= %d", snapshots.saves, afterConnect)
	}
	assertPersistentAgentUpdateAndDelete(t, ctx, store, principal, created.Agent.AgentID)
}

func TestPersistentAIAgentClientStoreFixtureCreateSaves(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	created, err := store.CreateAIAgentFromOnboardingFixture(
		ctx,
		principal,
		"riido_pm",
		CreateAgentConfigurationRequest{
			Name:       "fixture persistence agent",
			Visibility: AgentVisibilityPrivate,
			RuntimeID:  "runtime-cursor-dev",
			ModelID:    stringPtr("cursor-auto"),
		},
	)
	if err != nil || created.Agent.AgentID == "" {
		t.Fatalf("CreateAIAgentFromOnboardingFixture = %+v, %v", created, err)
	}
	if snapshots.saves == 0 {
		t.Fatalf("fixture create did not save snapshot")
	}
}
