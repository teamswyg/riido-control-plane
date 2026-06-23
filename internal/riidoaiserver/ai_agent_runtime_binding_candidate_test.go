package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestListAIAgentDaemonAgentBindingsAllowsRuntimeSharedAgents(t *testing.T) {
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
	if len(cursorBindings) != 2 {
		t.Fatalf("cursor binding count = %d, want 2: %+v", len(cursorBindings), cursorBindings)
	}
	seen := map[string]bool{}
	for _, binding := range cursorBindings {
		seen[binding.AgentID] = true
	}
	if !seen["agent-cursor-active"] || !seen["agent-cursor-idle"] {
		t.Fatalf("shared runtime bindings = %+v", cursorBindings)
	}
}
