package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentRuntimeAssignmentAllowsDuplicateCreate(t *testing.T) {
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

	second, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "영실 Claude",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent duplicate runtime: %v", err)
	}
	if second.Agent.AgentID == created.Agent.AgentID ||
		second.Agent.RuntimeID != created.Agent.RuntimeID {
		t.Fatalf("duplicate runtime create = first:%+v second:%+v", created.Agent, second.Agent)
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
