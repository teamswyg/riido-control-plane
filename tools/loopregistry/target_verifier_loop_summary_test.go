package main

import (
	"strings"
	"testing"
)

func TestTargetVerifierLoopSummaryUsesComponentLoops(t *testing.T) {
	got := targetVerifierLoopSummaryFor(&targetVerifierPlan{
		Components: []targetVerifierComponent{
			{LoopIDs: []string{"loop-b", "loop-a"}},
			{LoopIDs: []string{"loop-a", "loop-c"}},
		},
	}, "evidence.json")
	for _, want := range []string{
		"loops: loop-a, loop-b",
		"+1 more in evidence.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("loop summary missing %q: %s", want, got)
		}
	}
}

func TestTargetVerifierLoopSummarySkipsMissingPlan(t *testing.T) {
	if got := targetVerifierLoopSummaryFor(nil, "evidence.json"); got != "" {
		t.Fatalf("loop summary = %q", got)
	}
}
