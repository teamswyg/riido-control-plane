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
	if !strings.Contains(got, "필수 정보·작업 공간·권한") {
		t.Fatalf("no-workspace guidance missing: %q", got)
	}

	if !strings.Contains(got, "실제 작업을 수행하고 결과를 검증") || strings.Contains(got, "단계로 제시하며 마무리") {
		t.Fatalf("assignment must require execution when ready: %q", got)
	}

	// Empty instruction still yields the guidelines, so the system prompt is
	// never empty and the agent always has the Korean + graceful-failure rules.
	if augmentAgentInstruction("   ") != agentResponseGuidelines {
		t.Fatal("empty instruction should fall back to the shared guidelines")
	}
}
