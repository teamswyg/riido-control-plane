package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestAIAgentRuntimeAssignmentRejectsDuplicateCreate(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "리도 Claude",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent first: %v", err)
	}

	_, err = store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "영실 Claude",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	if !errors.Is(err, ErrAIAgentRuntimeAlreadyAssigned) {
		t.Fatalf("CreateAIAgent duplicate err=%v", err)
	}

	if _, err := store.DeleteAIAgent(ctx, principal, created.Agent.AgentID); err != nil {
		t.Fatalf("DeleteAIAgent first: %v", err)
	}
	if _, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "영실 Claude",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	}); err != nil {
		t.Fatalf("CreateAIAgent after delete: %v", err)
	}
}

func TestAIAgentRuntimeAssignmentRejectsUpdateToOccupiedRuntime(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "Cursor holder",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent: %v", err)
	}
	if _, err := store.UpdateAIAgentConfiguration(ctx, principal, created.Agent.AgentID, UpdateAgentConfigurationRequest{
		ModelID: stringPtr("cursor-fast"),
	}); err != nil {
		t.Fatalf("UpdateAIAgentConfiguration same runtime: %v", err)
	}

	_, err = store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{
		RuntimeID: "runtime-cursor-dev",
		ModelID:   stringPtr("cursor-auto"),
	})
	if !errors.Is(err, ErrAIAgentRuntimeAlreadyAssigned) {
		t.Fatalf("UpdateAIAgentConfiguration occupied runtime err=%v", err)
	}
}
