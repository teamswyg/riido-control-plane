package main

import "testing"

func TestLoopClosureAuditRejectsHarnessWithoutCandidateArtifact(t *testing.T) {
	m, deps := loadForTest(t)
	for i := range deps.registry.Loops {
		if deps.registry.Loops[i].ID == "provider_acceptance_harness" {
			deps.registry.Loops[i].Evidence = nil
		}
	}
	if err := verifyAll("../..", m, deps); err == nil {
		t.Fatal("expected missing candidate artifact to fail")
	}
}

func TestLoopClosureAuditEvidenceExposesHarnessSummaryProofSurface(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	proof := findProof(t, got, "harness_summary:closed_loop_candidate_harness_summary")
	if proof.Surface == nil || proof.Surface.HarnessCount < 3 ||
		len(proof.Surface.HarnessLoops) < 3 ||
		len(proof.Surface.HarnessCandidateArtifacts) < 3 {
		t.Fatalf("incomplete harness summary proof surface: %+v", proof.Surface)
	}
}
