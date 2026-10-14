package main

import "testing"

func TestLoopClosureAuditEvidenceHasNoProofSurfaceGaps(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	if got.ProofCount == 0 || got.ProofSurfaceCount != got.ProofCount {
		t.Fatalf("proof surface coverage = %d/%d",
			got.ProofSurfaceCount, got.ProofCount)
	}
	if got.ProofSurfaceGapCount != 0 {
		t.Fatalf("proof surface gaps = %d", got.ProofSurfaceGapCount)
	}
	for _, req := range got.Requirements {
		if req.ProofSurfaceCount != req.ProofCount {
			t.Fatalf("requirement %s proof surface coverage = %d/%d",
				req.ID, req.ProofSurfaceCount, req.ProofCount)
		}
	}
}
