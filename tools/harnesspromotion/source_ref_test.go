package main

import "testing"

func TestHarnessFailureCandidateCarriesSourceRef(t *testing.T) {
	source := promotionSource{
		ID: "smoke", HarnessLoop: "provider_acceptance_harness",
		SourceWorkflow:  ".github/workflows/smoke.yml",
		SummaryArtifact: "smoke-summary", CandidateArtifact: "smoke-candidates",
		PromotionTarget: "closed_loop_candidate", FailureStatuses: []string{"failure"},
	}
	summary := liveSummary{
		ID: "smoke", LiveStatus: "failure", GeneratedAt: "2026-06-24T00:00:00Z",
		ExpiresAt: "2026-06-25T00:00:00Z", Run: runRecord{ID: "run-1"},
		EvidenceClaims: []liveClaim{{ID: "broken", Status: "not_verified"}},
	}
	got := buildCandidateEvidence(source, summary)
	ref := got.Candidates[0].SourceRef
	if ref.SourceWorkflow != source.SourceWorkflow ||
		ref.SummaryArtifact != source.SummaryArtifact ||
		ref.CandidateArtifact != source.CandidateArtifact ||
		ref.SourceExpiresAt != summary.ExpiresAt ||
		ref.Run.ID != "run-1" {
		t.Fatalf("source_ref = %+v", ref)
	}
}
