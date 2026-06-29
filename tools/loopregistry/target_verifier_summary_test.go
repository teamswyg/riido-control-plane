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
			Paths: []targetVerifierPath{
				{Path: "tools/loopregistry/b.go"},
				{Path: "docs/30-architecture/loop-registry.riido.json"},
			},
			Components: []targetVerifierComponent{
				{
					Component:        "docs/30-architecture",
					LoopIDs:          []string{"ai_thread_history"},
					ClaimIDs:         []string{"claim-a"},
					EvidenceChainIDs: []string{"chain-a"},
				},
				{
					Component:        "tools/loopregistry",
					LoopIDs:          []string{"closed_loop_candidate"},
					ClaimIDs:         []string{"claim-b"},
					EvidenceChainIDs: []string{"chain-b"},
				},
			},
			VerifierCommands: []string{
				"go test ./tools/a -count=1",
				"go test ./tools/b -count=1",
				"go test ./tools/c -count=1",
			},
			EntrypointCommands: []string{
				"go test ./tools/b -count=1",
				"go test ./tools/a -count=1",
			},
			FocusedCommands: []string{
				"go test ./tools/focused -count=1",
			},
		},
	}, ".git/riido-loop-registry-precommit-evidence.json")
	for _, want := range []string{
		"3 changed paths, 2 matched paths, 0 exact, 0 component-routed, 2 components, 3 commands",
		"paths: docs/30-architecture/loop-registry.riido.json, tools/loopregistry/b.go",
		"components: docs/30-architecture, tools/loopregistry",
		"loops: ai_thread_history, closed_loop_candidate",
		"claims: claim-a, claim-b",
		"chains: chain-a, chain-b",
		"focused: go test ./tools/focused -count=1",
		"entrypoints: go test ./tools/b -count=1",
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
