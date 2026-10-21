package riidoaiserver

import (
	"errors"
	"testing"
)

func TestAIAgentDeletionAuthorizationReloadsPersistentOwnership(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, fixture)
	stale, _ := newMembershipDynamoReplica(t, fixture)
	principal := profileColorTestPrincipal()
	agentID := "agent-owned-claude"
	if err := stale.AuthorizeAIAgentDeletion(t.Context(), principal, agentID); err != nil {
		t.Fatalf("initial owner denied: %v", err)
	}
	// Simulate a durable record changed on another replica while the reader's
	// normal snapshot cadence still has the old owner cached.
	writer.mu.Lock()
	agent := writer.agents[agentID]
	agent.OwnerPrincipalID = "user-2"
	writer.agents[agentID] = agent
	writer.mu.Unlock()
	if err := writer.saveSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	if stale.agents[agentID].OwnerPrincipalID != principal.PrincipalID {
		t.Fatal("fixture did not retain the stale owner")
	}
	if err := stale.AuthorizeAIAgentDeletion(t.Context(), principal, agentID); !errors.Is(err, ErrAIAgentNotFound) {
		t.Fatalf("stale owner deletion authorization=%v, want not found", err)
	}
	principal.PrincipalID = "user-2"
	if err := stale.AuthorizeAIAgentDeletion(t.Context(), principal, agentID); err != nil {
		t.Fatalf("durable owner denied: %v", err)
	}
}
