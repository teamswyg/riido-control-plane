package main

import "testing"

func TestClaimCoverageCandidatesCarryStructuredSubject(t *testing.T) {
	source := candidateSource{
		ID:                    "loop-closure-audit",
		SourceWorkflow:        ".github/workflows/loop-closure-audit.yml",
		SummaryArtifact:       "loop-closure-audit-evidence",
		CandidateArtifact:     "loop-closure-audit-closed-loop-candidates",
		HarnessLoop:           "loop_closure_audit",
		PromotionTarget:       "closed_loop_candidate",
		RequiredNextArtifacts: []string{"claim_binding"},
	}
	candidates := claimCoverageCandidates(source, claimCoverageGapDeps(), "2026-06-24T12:00:00Z", "2026-06-25T12:00:00Z")
	if len(candidates) != 1 || !hasClaimCoverageSubject(candidates) {
		t.Fatalf("claim coverage candidates = %+v", candidates)
	}
}

func hasClaimCoverageSubject(candidates []closedLoopCandidate) bool {
	for _, candidate := range candidates {
		if candidate.Subject == nil {
			continue
		}
		if candidate.Subject.Kind == "claim_coverage_gap" &&
			candidate.Subject.ClaimID != "" &&
			candidate.Subject.Loop != "" &&
			len(candidate.Subject.MissingDimensions) > 0 {
			return true
		}
	}
	return false
}
