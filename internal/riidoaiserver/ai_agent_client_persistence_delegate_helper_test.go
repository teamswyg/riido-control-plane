package riidoaiserver

import (
	"context"
	"testing"
)

func assertPersistentAgentUpdateAndDelete(
	t *testing.T,
	ctx context.Context,
	store *PersistentAIAgentClientStore,
	principal AuthorizationResult,
	agentID string,
) {
	t.Helper()
	updated, err := store.UpdateAIAgentConfiguration(ctx, principal, agentID, UpdateAgentConfigurationRequest{
		Name:       "delegate persistence agent updated",
		Visibility: AgentVisibilityPublic,
		RuntimeID:  "runtime-cursor-dev",
	})
	if err != nil || updated.Agent.Name != "delegate persistence agent updated" {
		t.Fatalf("UpdateAIAgentConfiguration = %+v, %v", updated, err)
	}
	if _, err := store.DeleteAIAgent(ctx, principal, agentID); err != nil {
		t.Fatalf("DeleteAIAgent: %v", err)
	}
}
