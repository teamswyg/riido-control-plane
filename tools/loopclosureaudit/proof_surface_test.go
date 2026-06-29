package main

import "testing"

func TestLoopClosureAuditEvidenceExposesClaimProofSurface(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	proof := findProof(t, got, "claim:pre_commit_must_run_claim_binding_impact")
	if proof.Surface == nil {
		t.Fatalf("proof surface missing: %+v", proof)
	}
	if len(proof.Surface.Files) == 0 ||
		len(proof.Surface.Verifiers) == 0 ||
		len(proof.Surface.GeneratedDocs) == 0 ||
		proof.Surface.SemanticHash == "" {
		t.Fatalf("incomplete claim proof surface: %+v", proof.Surface)
	}
}

func TestLoopClosureAuditEvidenceExposesLoopProofSurface(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	proof := findProof(t, got, "loop:provider_acceptance_harness")
	if proof.Surface == nil || len(proof.Surface.Providers) == 0 ||
		len(proof.Surface.PromotesTo) == 0 {
		t.Fatalf("incomplete loop proof surface: %+v", proof.Surface)
	}
}
