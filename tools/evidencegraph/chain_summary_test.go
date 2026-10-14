package main

import "testing"

func TestEvidenceGraphEvidenceExposesCompiledChainSummaries(t *testing.T) {
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	result, err := verifyAll("../..", m)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	if got.ChainSummary.ChainCount != len(m.Chains) {
		t.Fatalf("chain count mismatch: %+v", got.ChainSummary)
	}
	if got.ChainSummary.CompleteChains != len(m.Chains) {
		t.Fatalf("complete chain count mismatch: %+v", got.ChainSummary)
	}
	if got.ChainSummary.NextLoopCount != len(got.NextLoopSummary) {
		t.Fatalf("next-loop count mismatch: %+v", got.ChainSummary)
	}
	if sumNextLoopChains(got.NextLoopSummary) != len(m.Chains) {
		t.Fatalf("next-loop summaries do not cover all chains: %+v", got.NextLoopSummary)
	}
}

func sumNextLoopChains(items []nextLoopSummary) int {
	total := 0
	for _, item := range items {
		total += item.Chains
	}
	return total
}
