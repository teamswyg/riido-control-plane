package main

import "testing"

func TestHarnessPromotionManifestVerifies(t *testing.T) {
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestHarnessFailurePromotesUnverifiedClaims(t *testing.T) {
	source := promotionSource{
		ID: "smoke", HarnessLoop: "provider_acceptance_harness",
		PromotionTarget: "closed_loop_candidate", FailureStatuses: []string{"failure"},
		RequiredNextArtifacts: []string{"claim_binding", "verifier", "ci_gate"},
	}
	summary := liveSummary{ID: "smoke", LiveStatus: "failure", EvidenceClaims: []liveClaim{
		{ID: "ok", Status: "verified"}, {ID: "broken", Summary: "broken claim", Status: "not_verified"},
	}}
	got := buildCandidateEvidence(source, summary)
	if got.CandidateCount != 1 || got.Candidates[0].ID != "smoke:broken" {
		t.Fatalf("candidates = %+v", got.Candidates)
	}
}
