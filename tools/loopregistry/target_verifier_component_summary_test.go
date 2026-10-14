package main

import (
	"strings"
	"testing"
)

func TestTargetVerifierAnnotationSummaryIncludesComponents(t *testing.T) {
	got := targetVerifierAnnotationSummary(&targetVerifierPlan{
		Paths: []targetVerifierPath{
			{Path: "tools/loopregistry/b.go"},
			{Path: "docs/30-architecture/loop-registry.riido.json"},
			{Path: "tools/loopregistry/a.go"},
		},
		Components: []targetVerifierComponent{
			{
				Component:        "docs/30-architecture",
				LoopIDs:          []string{"loop-b"},
				ClaimIDs:         []string{"claim-b"},
				EvidenceChainIDs: []string{"chain-b"},
			},
			{
				Component:        "internal/riidoaiserver",
				LoopIDs:          []string{"loop-a"},
				ClaimIDs:         []string{"claim-a"},
				EvidenceChainIDs: []string{"chain-a"},
			},
			{
				Component:        "tools/loopregistry",
				LoopIDs:          []string{"loop-c"},
				ClaimIDs:         []string{"claim-c"},
				EvidenceChainIDs: []string{"chain-c"},
			},
		},
		CommandCount: 3,
		VerifierCommands: []string{
			"go test ./tools/a -count=1",
			"go test ./tools/b -count=1",
			"go test ./tools/c -count=1",
		},
	})
	for _, want := range []string{
		"paths: docs/30-architecture/loop-registry.riido.json, tools/loopregistry/a.go",
		"components: docs/30-architecture, internal/riidoaiserver",
		"+1 more in loop-registry-evidence",
		"loops: loop-a, loop-b",
		"claims: claim-a, claim-b",
		"chains: chain-a, chain-b",
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
