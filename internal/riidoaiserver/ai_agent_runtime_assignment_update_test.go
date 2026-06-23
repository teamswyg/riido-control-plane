package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentRuntimeAssignmentAllowsUpdateToOccupiedRuntime(t *testing.T) {
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

	updated, err := store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{
		RuntimeID: "runtime-cursor-dev",
		ModelID:   stringPtr("cursor-auto"),
	})
	if err != nil {
		t.Fatalf("UpdateAIAgentConfiguration occupied runtime: %v", err)
	}
	if updated.Agent.RuntimeID != created.Agent.RuntimeID {
		t.Fatalf("updated runtime = %q want %q", updated.Agent.RuntimeID, created.Agent.RuntimeID)
	}
}
