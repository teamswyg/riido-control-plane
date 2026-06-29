package main

import (
	"strings"
	"testing"
)

func TestTargetVerifierSummaryUsesImpactPlan(t *testing.T) {
	got := targetVerifierSummary(&impactEvidence{
		TargetVerifierPlan: &targetVerifierPlan{
			ChangedPathCount: 3,
			MatchedPathCount: 2,
			ComponentCount:   2,
			CommandCount:     3,
			Components: []targetVerifierComponent{
				{
					Component: "docs/30-architecture",
					LoopIDs:   []string{"ai_thread_history"},
					ClaimIDs:  []string{"claim-a"},
				},
				{
					Component: "tools/loopregistry",
					LoopIDs:   []string{"closed_loop_candidate"},
					ClaimIDs:  []string{"claim-b"},
				},
			},
			VerifierCommands: []string{
				"go test ./tools/a -count=1",
				"go test ./tools/b -count=1",
				"go test ./tools/c -count=1",
			},
		},
	}, ".git/riido-loop-registry-precommit-evidence.json")
	for _, want := range []string{
		"3 changed paths, 2 matched paths, 2 components, 3 commands",
		"components: docs/30-architecture, tools/loopregistry",
		"loops: ai_thread_history, closed_loop_candidate",
		"claims: claim-a, claim-b",
		"go test ./tools/a -count=1",
		"go test ./tools/b -count=1",
		"+1 more in .git/riido-loop-registry-precommit-evidence.json",
		"full_plan: .git/riido-loop-registry-precommit-evidence.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("summary missing %q: %s", want, got)
		}
	}
}

func TestTargetVerifierSummarySkipsMissingPlan(t *testing.T) {
	if got := targetVerifierSummary(&impactEvidence{}, "evidence.json"); got != "" {
		t.Fatalf("summary = %q", got)
	}
}
