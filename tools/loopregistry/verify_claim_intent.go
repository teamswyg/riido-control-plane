package main

import (
	"fmt"
	"os"
	"strings"
)

const verifierIntentClaimID = "loop_verifiers_must_accept_verify_intent"

func verifyVerifierIntentClaim(root string, tests map[string][]string, claim claimBinding) error {
	mainFiles := verifierIntentMainFiles(claim)
	if len(mainFiles) == 0 {
		return fmt.Errorf("claim %s must bind verifier main files", claim.ID)
	}
	for _, path := range mainFiles {
		if err := verifyAliasInMain(root, claim.ID, path); err != nil {
			return err
		}
	}
	if countVerifyAliasTests(claim, tests) < len(mainFiles) {
		return fmt.Errorf("claim %s must bind a VerifyAlias test per verifier main", claim.ID)
	}
	return nil
}

func verifierIntentMainFiles(claim claimBinding) []string {
	out := []string{}
	for _, path := range claim.Files {
		if strings.HasPrefix(path, "tools/") && strings.HasSuffix(path, "/main.go") {
			out = append(out, path)
		}
	}
	return out
}

func verifyAliasInMain(root, claimID, path string) error {
	data, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return err
	}
	text := string(data)
	if !strings.Contains(text, "\"verify\"") ||
		!strings.Contains(text, "alias for -check-doc") {
		return fmt.Errorf("claim %s file %s must define -verify alias", claimID, path)
	}
	return nil
}

func countVerifyAliasTests(claim claimBinding, tests map[string][]string) int {
	count := 0
	for _, verifier := range claim.Verifiers {
		if strings.Contains(verifier, "VerifyAlias") && len(tests[verifier]) > 0 {
			count++
		}
	}
	return count
}
