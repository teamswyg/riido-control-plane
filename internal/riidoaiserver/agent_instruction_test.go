package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAugmentAgentInstruction(t *testing.T) {
	// Guidelines are appended to a non-empty instruction.
	got := augmentAgentInstruction("You are a backend agent.")
	if !strings.Contains(got, "You are a backend agent.") {
		t.Fatalf("agent instruction dropped: %q", got)
	}
	if !strings.Contains(got, "한국어") {
		t.Fatalf("Korean response guideline missing: %q", got)
	}
	if !strings.Contains(got, "어느 경로") {
		t.Fatalf("no-workspace guidance missing: %q", got)
	}

	// Empty instruction still yields the guidelines, so the system prompt is
	// never empty and the agent always has the Korean + graceful-failure rules.
	if augmentAgentInstruction("   ") != agentResponseGuidelines {
		t.Fatal("empty instruction should fall back to the shared guidelines")
	}
}
