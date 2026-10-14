package main

import "testing"

func TestClaimRuleRegistryCoversVerifierIntentClaim(t *testing.T) {
	for _, rule := range claimVerificationRules {
		if rule.id == verifierIntentClaimID {
			return
		}
	}
	t.Fatalf("missing claim rule for %s", verifierIntentClaimID)
}

func TestClaimRuleRegistryRejectsDuplicateIDs(t *testing.T) {
	rules := []claimVerificationRule{
		{id: verifierIntentClaimID, verify: verifyVerifierIntentClaim},
		{id: verifierIntentClaimID, verify: verifyVerifierIntentClaim},
	}
	if err := verifyClaimRuleRegistry(rules); err == nil {
		t.Fatal("expected duplicate claim rule to fail")
	}
}
