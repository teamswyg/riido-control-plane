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
	if proof.Surface == nil || len(proof.Surface.Observes) == 0 ||
		len(proof.Surface.PromotesTo) == 0 ||
		!contains(proof.Surface.Observes, "visible_runtime_acceptance") {
		t.Fatalf("incomplete loop proof surface: %+v", proof.Surface)
	}
}

func TestLoopClosureAuditEvidenceExposesGraphEdgeProofSurface(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	proof := findProof(t, got,
		"graph_edge:provider_acceptance_harness:promotes_failure_to:closed_loop_candidate")
	if proof.Surface == nil || proof.Surface.From == "" ||
		proof.Surface.To == "" || proof.Surface.Relation == "" {
		t.Fatalf("incomplete graph edge proof surface: %+v", proof.Surface)
	}
}

func TestLoopClosureAuditEvidenceExposesGraphSummaryProofSurface(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	proof := findProof(t, got, "graph_summary:evidence_graph_chain_summary")
	if proof.Surface == nil || proof.Surface.GraphChainCount == 0 ||
		proof.Surface.GraphCompleteChains != proof.Surface.GraphChainCount ||
		proof.Surface.GraphNextLoopCount == 0 ||
		len(proof.Surface.GraphNextLoops) == 0 {
		t.Fatalf("incomplete graph summary proof surface: %+v", proof.Surface)
	}
}

func TestLoopClosureAuditEvidenceExposesPreCommitHookProofSurface(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	proof := findProof(t, got, "pre_commit_hook:loop-registry-claim-binding")
	if proof.Surface == nil || proof.Surface.PreCommitHook == "" ||
		len(proof.Surface.Contains) == 0 {
		t.Fatalf("incomplete pre-commit hook proof surface: %+v", proof.Surface)
	}
}
