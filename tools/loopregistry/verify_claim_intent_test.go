package main

import "testing"

func TestVerifierIntentClaimRejectsMainWithoutVerifyAlias(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	claim := claimForTest(t, m, verifierIntentClaimID)
	claim.Files = []string{"tools/providerstatus/main.go"}
	if err := verifyClaim("../..", loopIDsForTest(m), map[string][]string{
		"TestProviderStatusVerifyAlias": {"tools/providerstatus"},
	}, hashes, claim); err == nil {
		t.Fatal("expected missing -verify alias to fail")
	}
}

func TestVerifierIntentClaimRejectsMissingAliasTest(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	claim := claimForTest(t, m, verifierIntentClaimID)
	if err := verifyClaim("../..", loopIDsForTest(m), map[string][]string{}, hashes, claim); err == nil {
		t.Fatal("expected missing VerifyAlias tests to fail")
	}
}

func claimForTest(t *testing.T, m manifest, id string) claimBinding {
	t.Helper()
	for _, claim := range m.Claims {
		if claim.ID == id {
			return claim
		}
	}
	t.Fatalf("missing claim %s", id)
	return claimBinding{}
}
