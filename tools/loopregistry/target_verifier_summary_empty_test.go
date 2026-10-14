package main

import "testing"

func TestTargetVerifierSummarySkipsMissingPlan(t *testing.T) {
	if got := targetVerifierSummary(&impactEvidence{}, "evidence.json"); got != "" {
		t.Fatalf("summary = %q", got)
	}
}
