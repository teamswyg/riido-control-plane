package main

import "testing"

func TestHarnessCandidateCarriesAdoptionPlan(t *testing.T) {
	source := promotionSource{
		ID: "smoke", HarnessLoop: "provider_acceptance_harness",
		PromotionTarget: "closed_loop_candidate", FailureStatuses: []string{"failure"},
		SummaryPath: "out/smoke-summary.json", CandidatePath: "out/smoke-candidates.json",
		RequiredNextArtifacts: []string{"claim_binding", "decision_record"},
	}
	got := buildCandidateEvidence(source, liveSummary{
		ID: "smoke", LiveStatus: "failure",
		EvidenceClaims: []liveClaim{{ID: "broken", Status: "not_verified"}},
	})
	plan := got.Candidates[0].AdoptionPlan
	if len(plan) != 2 || plan[0].Artifact != "claim_binding" {
		t.Fatalf("adoption plan = %+v", plan)
	}
	if plan[1].Command == "" || plan[1].Artifact != "decision_record" {
		t.Fatalf("decision adoption step = %+v", plan[1])
	}
}
