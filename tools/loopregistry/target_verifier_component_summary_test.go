package main

import (
	"strings"
	"testing"
)

func TestTargetVerifierAnnotationSummaryIncludesComponents(t *testing.T) {
	got := targetVerifierAnnotationSummary(&targetVerifierPlan{
		Components: []targetVerifierComponent{
			{Component: "docs/30-architecture", LoopIDs: []string{"loop-b"}},
			{Component: "internal/riidoaiserver", LoopIDs: []string{"loop-a"}},
			{Component: "tools/loopregistry", LoopIDs: []string{"loop-c"}},
		},
		CommandCount: 3,
		VerifierCommands: []string{
			"go test ./tools/a -count=1",
			"go test ./tools/b -count=1",
			"go test ./tools/c -count=1",
		},
	})
	for _, want := range []string{
		"components: docs/30-architecture, internal/riidoaiserver",
		"+1 more in loop-registry-evidence",
		"loops: loop-a, loop-b",
		"commands: go test ./tools/a -count=1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation summary missing %q: %s", want, got)
		}
	}
}

func TestTargetVerifierComponentSummarySkipsMissingPlan(t *testing.T) {
	if got := targetVerifierComponentSummaryFor(nil, "evidence.json"); got != "" {
		t.Fatalf("component summary = %q", got)
	}
}
