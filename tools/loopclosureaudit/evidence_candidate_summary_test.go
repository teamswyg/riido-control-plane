package main

import "testing"

func TestLoopClosureAuditEvidenceExposesCandidateSummaryIDs(t *testing.T) {
	m := manifest{
		Sources: []candidateSource{{
			ID:                "loop-closure-audit",
			CandidateArtifact: "loop-closure-audit-closed-loop-candidates",
			PromotionTarget:   "closed_loop_candidate",
		}},
		ResidualGaps: []residualGap{{ID: "residual-a"}},
	}
	got := newEvidence(m, claimCoverageGapDeps())
	summary := got.CandidateSummary
	if summary.CandidateCount != got.CandidateCount {
		t.Fatalf("candidate summary count = %d, want %d", summary.CandidateCount, got.CandidateCount)
	}
	if !hasString(summary.CandidateIDs, "loop-closure-audit:residual-a") {
		t.Fatalf("summary missing residual candidate id: %+v", summary)
	}
	if !hasString(summary.CandidateIDs, "loop-closure-audit:claim_coverage:claim-a") {
		t.Fatalf("summary missing claim coverage candidate id: %+v", summary)
	}
}

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
