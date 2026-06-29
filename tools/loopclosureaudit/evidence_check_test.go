package main

import "testing"

func TestLoopClosureAuditEvidenceExposesCheckDetails(t *testing.T) {
	got := newEvidence(manifest{
		Requirements: []requirement{{
			ID:        "r1",
			Statement: "require proof",
			Checks: []check{{
				Kind:     "workflow",
				Path:     ".github/workflows/example.yml",
				Contains: []string{"evidence"},
			}},
		}},
	})
	requirement := got.Requirements[0]
	checks := requirement.Checks
	if len(checks) != 1 || checks[0].Path == "" || len(checks[0].Contains) == 0 {
		t.Fatalf("expected check details in evidence, got %+v", got.Requirements[0])
	}
	if requirement.Status != "verified" || requirement.ProofCount != 1 {
		t.Fatalf("expected verified proof summary, got %+v", requirement)
	}
	proofs := requirement.Proofs
	if len(proofs) != 1 || proofs[0].Status != "verified" ||
		proofs[0].Key == "" {
		t.Fatalf("expected proof keys in evidence, got %+v", requirement)
	}
}
