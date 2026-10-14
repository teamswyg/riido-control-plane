package main

import (
	"strings"
	"testing"
)

func TestTargetVerifierClaimSummaryUsesComponentClaims(t *testing.T) {
	got := targetVerifierClaimSummaryFor(&targetVerifierPlan{
		Components: []targetVerifierComponent{
			{ClaimIDs: []string{"claim-b", "claim-a"}},
			{ClaimIDs: []string{"claim-a", "claim-c"}},
		},
	}, "evidence.json")
	for _, want := range []string{
		"claims: claim-a, claim-b",
		"+1 more in evidence.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("claim summary missing %q: %s", want, got)
		}
	}
}

func TestTargetVerifierClaimSummarySkipsMissingPlan(t *testing.T) {
	if got := targetVerifierClaimSummaryFor(nil, "evidence.json"); got != "" {
		t.Fatalf("claim summary = %q", got)
	}
}
