package main

import (
	"strings"
	"testing"
)

func TestLoopRefreshDispatchWorkflowDoesNotEvalCommands(t *testing.T) {
	text := loopRefreshDispatchWorkflowText(t)
	for _, forbidden := range []string{"eval ", ".commands[].command", "bash -c"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("workflow must not execute raw command strings: found %q", forbidden)
		}
	}
	if !strings.Contains(text, ".workflow_file") ||
		!strings.Contains(text, `gh workflow run "$workflow" --ref main`) {
		t.Fatalf("workflow must dispatch only structured workflow_file values")
	}
}

func TestLoopRefreshDispatchWorkflowConsumesDecisionCommands(t *testing.T) {
	text := loopRefreshDispatchWorkflowText(t)
	if !strings.Contains(text, "closed-loop-candidate-decision-commands") {
		t.Fatal("workflow must download candidate decision command evidence")
	}
	if !strings.Contains(text, "-commands-in \"$decision_commands\"") {
		t.Fatal("workflow must pass candidate decision commands into dispatcher")
	}
	if !strings.Contains(text, "go run ./tools/closedloopcandidatedecision") ||
		!strings.Contains(text, "-commands-out out/sample/sample-decision-refresh-commands.json") {
		t.Fatal("workflow sample must use the real candidate decision command producer")
	}
	if strings.Count(text, "-commands-in") < 4 {
		t.Fatal("workflow must cover repeated command inputs in sample and live runs")
	}
}

func TestLoopRefreshDispatchWorkflowUsesCommandFixture(t *testing.T) {
	text := loopRefreshDispatchWorkflowText(t)
	if !strings.Contains(text, "loop-refresh-commands.fixture.json") {
		t.Fatal("workflow sample must consume the command fixture")
	}
	if strings.Contains(text, "cat > out/sample-loop-refresh-commands.json") {
		t.Fatal("workflow must not embed loop refresh command JSON inline")
	}
}
