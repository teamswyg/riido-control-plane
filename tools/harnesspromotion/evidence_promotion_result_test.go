package main

import "testing"

func TestHarnessPromotionEvidenceCarriesPromotionResult(t *testing.T) {
	promotion := &promotionResult{
		ID:                "smoke",
		CandidateArtifact: "out/candidates.json",
		LiveStatus:        "failure",
		CandidateCount:    1,
		CandidateIDs:      []string{"smoke:broken"},
	}
	got := newEvidence(manifest{}, verifyResult{}, promotion)
	if got.PromotionResult == nil ||
		got.PromotionResult.CandidateArtifact != "out/candidates.json" ||
		got.PromotionResult.CandidateIDs[0] != "smoke:broken" {
		t.Fatalf("promotion result missing from evidence: %+v", got.PromotionResult)
	}
}
