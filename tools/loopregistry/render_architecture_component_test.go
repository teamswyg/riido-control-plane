package main

import (
	"strings"
	"testing"
)

func TestRenderArchitectureComponentsIncludesLoopSample(t *testing.T) {
	var b strings.Builder
	renderArchitectureComponents(&b, []architectureComponent{{
		Component:        "tools/example",
		PathCount:        1,
		LoopIDs:          []string{"closed_loop_candidate"},
		ClaimIDs:         []string{"claim-a"},
		VerifierCommands: []string{"go test ./tools/example -count=1"},
		EvidenceChainIDs: []string{"chain-a"},
	}})
	got := b.String()
	for _, want := range []string{"Loop sample", "`closed_loop_candidate`"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered architecture components missing %q:\n%s", want, got)
		}
	}
}
