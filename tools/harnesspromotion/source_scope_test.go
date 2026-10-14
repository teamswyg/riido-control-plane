package main

import "testing"

func TestHarnessPromotionEvidenceSeparatesSourceScopes(t *testing.T) {
	m := loadPromotionManifestForTest(t)
	result, err := verifyAll("../..", m)
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceCount != result.SidecarSourceCount {
		t.Fatalf("sources = %d, sidecar = %d", result.SourceCount, result.SidecarSourceCount)
	}
	if result.LoopOwnedCandidateProducerCount == 0 {
		t.Fatal("expected loop-owned candidate producer count")
	}
	got := newEvidence(m, result, nil)
	if got.LoopOwnedCandidateProducerCount != result.LoopOwnedCandidateProducerCount {
		t.Fatalf("evidence loop-owned producers = %d", got.LoopOwnedCandidateProducerCount)
	}
}
