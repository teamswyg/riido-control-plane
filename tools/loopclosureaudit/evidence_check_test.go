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
	checks := got.Requirements[0].Checks
	if len(checks) != 1 || checks[0].Path == "" || len(checks[0].Contains) == 0 {
		t.Fatalf("expected check details in evidence, got %+v", got.Requirements[0])
	}
}
