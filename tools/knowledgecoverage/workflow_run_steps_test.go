package main

import "testing"

func TestWorkflowTextRunsEvidenceToolRequiresSameRunStep(t *testing.T) {
	text := "steps:\n- run: go run ./tools/example -check-doc\n- run: echo -evidence-out out/example.json\n"
	if workflowTextRunsEvidenceTool(text, "./tools/example") {
		t.Fatal("evidence flags must not be borrowed from another run step")
	}
}

func TestWorkflowTextRunsEvidenceToolMatchesBlockRun(t *testing.T) {
	text := "steps:\n- run: |\n    go run ./tools/example \\\n      -check-doc \\\n      -evidence-out out/example.json\n"
	if !workflowTextRunsEvidenceTool(text, "./tools/example") {
		t.Fatal("expected block run evidence command match")
	}
}

func TestWorkflowTextRunsToolIgnoresNonRunMentions(t *testing.T) {
	text := "paths:\n  - tools/example/**\nsteps:\n- run: echo no-op\n"
	if workflowTextRunsTool(text, "./tools/example") {
		t.Fatal("tool mentions outside run steps must not count as executable evidence")
	}
}
