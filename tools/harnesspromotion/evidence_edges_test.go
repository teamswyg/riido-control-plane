package main

import "testing"

func TestHarnessPromotionEvidenceExposesPromotionEdges(t *testing.T) {
	m := manifest{Sources: []promotionSource{
		{HarnessLoop: "provider_acceptance_harness", PromotionTarget: "closed_loop_candidate"},
		{HarnessLoop: "provider_acceptance_harness", PromotionTarget: "closed_loop_candidate"},
		{HarnessLoop: "ai_agent_load_harness", PromotionTarget: "closed_loop_candidate"},
	}}
	got := newEvidence(m, verifyResult{}, nil)
	if got.PromotionEdgeCount != 2 || len(got.PromotionEdges) != 2 {
		t.Fatalf("promotion edges = %+v", got.PromotionEdges)
	}
	first := got.PromotionEdges[0]
	if first.From != "ai_agent_load_harness" ||
		first.To != "closed_loop_candidate" ||
		first.Relation != "promotes_failure_to" {
		t.Fatalf("first promotion edge = %+v", first)
	}
}
