package main

import "testing"

func TestAIThreadHistoryClaimsHaveCoverageTokens(t *testing.T) {
	requireNoClaimCoverageGapForLoop(t, "ai_thread_history")
}

func TestRepairedClaimsHaveCoverageTokens(t *testing.T) {
	for _, claim := range []string{
		"harness_promotion_must_run_after_failure",
		"claim_surface_evidence_must_expose_code_test_doc_binding",
		"candidate_decision_next_artifact_must_be_required",
		"claim_bound_file_changes_require_reasoning_chain",
		"claim_bound_paths_must_trigger_loop_registry",
		"claim_impact_evidence_must_expose_changed_files",
		"claim_meaning_changes_require_code_or_test_surface",
		"claim_meaning_changes_require_reasoning_chain",
		"claim_verifier_commands_must_surface_as_ci_annotations",
		"closed_loop_candidate_consumers_must_reject_expired_candidates",
		"closed_loop_candidate_evidence_must_self_expire",
		"closed_loop_candidates_must_carry_adoption_plan",
		"closed_loop_candidates_must_carry_promotion_edge",
		"closed_loop_candidates_must_carry_source_ref",
		"evidence_graph_chain_changes_require_executable_surface",
		"evidence_graph_evidence_must_expose_full_chain",
		"evidence_graph_must_cover_loop_registry_claims",
		"evidence_graph_refs_must_trigger_evidence_workflow",
		"expiring_loops_must_schedule_refresh",
		"expired_loop_evidence_must_select_refresh_commands",
		"expired_loop_refresh_commands_must_dispatch_safe_workflows",
		"harness_like_workflows_must_be_registered_or_excluded",
		"loop_evidence_artifacts_must_have_refresh_owners",
		"loop_evidence_artifacts_must_self_expire",
		"loop_evidence_sources_must_be_claim_covered",
		"loop_failure_conditions_must_be_claim_covered",
		"loop_observation_tokens_must_be_claim_covered",
		"loop_registry_evidence_must_expose_graph_edges",
		"loop_verify_tokens_must_be_claim_covered",
		"pre_commit_must_run_claim_binding_impact",
	} {
		t.Run(claim, func(t *testing.T) {
			requireNoClaimCoverageGapForClaim(t, claim)
		})
	}
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
