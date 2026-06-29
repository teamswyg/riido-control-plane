package main

import (
	"strings"
	"testing"
)

func TestTargetVerifierChainSummaryUsesComponentChains(t *testing.T) {
	got := targetVerifierChainSummaryFor(&targetVerifierPlan{
		Components: []targetVerifierComponent{
			{EvidenceChainIDs: []string{"chain-b", "chain-a"}},
			{EvidenceChainIDs: []string{"chain-c", "chain-a"}},
		},
	}, "evidence.json")
	for _, want := range []string{
		"chains: chain-a, chain-b",
		"+1 more in evidence.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("chain summary missing %q: %s", want, got)
		}
	}
}

func TestTargetVerifierChainSummarySkipsMissingPlan(t *testing.T) {
	if got := targetVerifierChainSummaryFor(nil, "evidence.json"); got != "" {
		t.Fatalf("chain summary = %q", got)
	}
}
