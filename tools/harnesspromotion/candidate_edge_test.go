package main

import "testing"

func TestHarnessFailureCandidateCarriesPromotionEdge(t *testing.T) {
	source := promotionSource{
		ID:              "smoke",
		HarnessLoop:     "provider_acceptance_harness",
		PromotionTarget: "closed_loop_candidate",
		FailureStatuses: []string{"failure"},
	}
	summary := liveSummary{
		ID:         "smoke",
		LiveStatus: "failure",
		EvidenceClaims: []liveClaim{{
			ID:      "provider_smoke",
			Summary: "provider smoke failed",
			Status:  "not_verified",
		}},
	}
	got := buildCandidateEvidence(source, summary)
	if got.CandidateCount != 1 {
		t.Fatalf("candidate count = %d", got.CandidateCount)
	}
	edge := got.Candidates[0].PromotionEdge
	if edge.From != source.HarnessLoop ||
		edge.To != source.PromotionTarget ||
		edge.Relation != "promotes_failure_to" {
		t.Fatalf("promotion edge = %+v", edge)
	}
}
