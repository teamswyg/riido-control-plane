package main

import "fmt"

type claimVerificationRule struct {
	id     string
	verify func(string, map[string][]string, claimBinding) error
}

var claimVerificationRules = []claimVerificationRule{
	{id: verifierIntentClaimID, verify: verifyVerifierIntentClaim},
}

func verifyClaimRules(root string, tests map[string][]string, claim claimBinding) error {
	for _, rule := range claimVerificationRules {
		if rule.id == claim.ID {
			return rule.verify(root, tests, claim)
		}
	}
	return nil
}

func verifyClaimRuleRegistry(rules []claimVerificationRule) error {
	seen := map[string]bool{}
	for _, rule := range rules {
		if rule.id == "" || rule.verify == nil {
			return fmt.Errorf("claim rule must bind id and verifier")
		}
		if seen[rule.id] {
			return fmt.Errorf("duplicate claim rule %s", rule.id)
		}
		seen[rule.id] = true
	}
	return nil
}
