package main

import "testing"

func TestLoopClosureAuditEvidenceExposesCandidateSurface(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	want := len(m.ResidualGaps) + len(claimCoverageGaps(deps))
	if got.CandidateArtifact != "loop-closure-audit-closed-loop-candidates" {
		t.Fatalf("candidate artifact = %q", got.CandidateArtifact)
	}
	if got.CandidateSourceID != "loop-closure-audit" {
		t.Fatalf("candidate source id = %q", got.CandidateSourceID)
	}
	if got.CandidateTarget != "closed_loop_candidate" {
		t.Fatalf("candidate target = %q", got.CandidateTarget)
	}
	if got.CandidateCount != want {
		t.Fatalf("candidate count = %d, want %d", got.CandidateCount, want)
	}
}
