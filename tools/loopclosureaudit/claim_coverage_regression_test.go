package main

import "testing"

func TestAIThreadHistoryClaimsHaveCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForLoop(t, "ai_thread_history")
}

func TestHarnessPromotionClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "harness_promotion_must_run_after_failure")
}

func TestClaimSurfaceEvidenceClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "claim_surface_evidence_must_expose_code_test_doc_binding")
}

func TestCandidateDecisionNextArtifactClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "candidate_decision_next_artifact_must_be_required")
}

func TestClaimBoundFileReasoningChainClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "claim_bound_file_changes_require_reasoning_chain")
}

func TestClaimBoundPathsTriggerClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "claim_bound_paths_must_trigger_loop_registry")
}

func TestClaimImpactChangedFilesClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "claim_impact_evidence_must_expose_changed_files")
}

func TestClaimMeaningCodeSurfaceClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "claim_meaning_changes_require_code_or_test_surface")
}

func TestClaimMeaningReasoningChainClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "claim_meaning_changes_require_reasoning_chain")
}

func TestClaimVerifierAnnotationClaimHasCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForClaim(t, "claim_verifier_commands_must_surface_as_ci_annotations")
}

func requireNoClaimCoverageGapForLoop(t *testing.T, loop string) {
	t.Helper()
	_, deps := loadForTest(t)
	for _, gap := range claimCoverageGaps(deps) {
		if gap.Loop == loop {
			t.Fatalf("%s claim coverage gap remains: %+v", loop, gap)
		}
	}
}

func requireNoClaimCoverageGapForClaim(t *testing.T, claim string) {
	t.Helper()
	_, deps := loadForTest(t)
	for _, gap := range claimCoverageGaps(deps) {
		if gap.ClaimID == claim {
			t.Fatalf("%s coverage gap remains: %+v", claim, gap)
		}
	}
}
