package main

import "testing"

func TestLoopClosureAuditCoversExecutableLoopRegistrySSOT(t *testing.T) {
	m, _ := loadForTest(t)
	req, ok := findRequirement(m.Requirements, "loop_registry_ssot_defines_executable_loop_manifests")
	if !ok {
		t.Fatal("loop registry executable SSOT requirement is missing")
	}
	wantClaims := []string{
		"loop_observation_tokens_must_be_claim_covered",
		"loop_verify_tokens_must_be_claim_covered",
		"loop_failure_conditions_must_be_claim_covered",
		"loop_evidence_sources_must_be_claim_covered",
		"expiring_loops_must_schedule_refresh",
		"loop_registry_evidence_must_expose_refresh_plans",
	}
	for _, claim := range wantClaims {
		if !requirementHasClaim(req, claim) {
			t.Fatalf("requirement %s missing claim %s", req.ID, claim)
		}
	}
}

func findRequirement(requirements []requirement, id string) (requirement, bool) {
	for _, req := range requirements {
		if req.ID == id {
			return req, true
		}
	}
	return requirement{}, false
}

func requirementHasClaim(req requirement, claimID string) bool {
	for _, check := range req.Checks {
		if check.Kind == "claim" && check.ID == claimID {
			return true
		}
	}
	return false
}
