package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestListAIAgentDaemonAgentBindingsDedupesLegacyRuntimeAgents(t *testing.T) {
	now := time.Now().UTC()
	store := NewDevelopmentAIAgentClientStore()
	prepareRuntimeBindingCandidateFixture(store, now)

	resp, err := store.ListAIAgentDaemonAgentBindings(
		context.Background(),
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID},
		"device-dev-macbook",
	)
	if err != nil {
		t.Fatalf("ListAIAgentDaemonAgentBindings: %v", err)
	}

	var cursorBindings []AgentRuntimeBinding
	for _, binding := range resp.Bindings {
		if binding.RuntimeID == "runtime-cursor-dev" {
			cursorBindings = append(cursorBindings, binding)
		}
	}
	if len(cursorBindings) != 1 {
		t.Fatalf("cursor binding count = %d, want 1: %+v", len(cursorBindings), cursorBindings)
	}
	if cursorBindings[0].AgentID != "agent-cursor-active" {
		t.Fatalf("selected binding = %+v, want active agent", cursorBindings[0])
	}
}
