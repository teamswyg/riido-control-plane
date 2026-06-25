package main

import "testing"

func TestClaimCoverageGapsReportMissingDimensions(t *testing.T) {
	deps := dependencies{registry: loopRegistry{
		Loops: []registryLoop{{
			ID:        "loop-a",
			Observes:  []string{"observation"},
			Verifies:  []string{"verification"},
			FailsWhen: []string{"failure"},
		}},
		Claims: []registryClaim{{ID: "claim-a", Loop: "loop-a"}},
	}}
	gaps := claimCoverageGaps(deps)
	if len(gaps) != 1 || gaps[0].ClaimID != "claim-a" {
		t.Fatalf("gaps = %+v", gaps)
	}
	if got := gaps[0].MissingDimensions; len(got) != 3 {
		t.Fatalf("missing dimensions = %+v", got)
	}
}

func TestLoopClosureAuditEvidenceExposesClaimCoverageGaps(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	if got.ClaimCoverageGapCount == 0 || len(got.ClaimCoverageGaps) == 0 {
		t.Fatalf("expected claim coverage gaps in evidence")
	}
	first := got.ClaimCoverageGaps[0]
	if first.ClaimID == "" || first.Loop == "" || len(first.MissingDimensions) == 0 {
		t.Fatalf("incomplete claim coverage gap = %+v", first)
	}
}

func TestAIThreadHistoryClaimsHaveCoverageTokens(t *testing.T) {
	_, deps := loadForTest(t)
	for _, gap := range claimCoverageGaps(deps) {
		if gap.Loop == "ai_thread_history" {
			t.Fatalf("ai_thread_history claim coverage gap remains: %+v", gap)
		}
	}
}

func TestHarnessPromotionClaimHasCoverageTokens(t *testing.T) {
	_, deps := loadForTest(t)
	for _, gap := range claimCoverageGaps(deps) {
		if gap.ClaimID == "harness_promotion_must_run_after_failure" {
			t.Fatalf("harness promotion claim coverage gap remains: %+v", gap)
		}
	}
}

func TestClaimSurfaceEvidenceClaimHasCoverageTokens(t *testing.T) {
	_, deps := loadForTest(t)
	for _, gap := range claimCoverageGaps(deps) {
		if gap.ClaimID == "claim_surface_evidence_must_expose_code_test_doc_binding" {
			t.Fatalf("claim surface evidence coverage gap remains: %+v", gap)
		}
	}
}
