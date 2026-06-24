package main

import "fmt"

func verifyChain(
	root string,
	c chain,
	seen map[string]bool,
	result *verifyResult,
	registry loopRegistryIndex,
	coveredClaims map[string]bool,
) error {
	if c.ID == "" || c.Observation == "" || c.Hypothesis == "" {
		return fmt.Errorf("chain id, observation, and hypothesis are required")
	}
	if c.Decision == "" || c.NextLoop == "" {
		return fmt.Errorf("%s decision and next_loop are required", c.ID)
	}
	if !registry.Loops[c.NextLoop] {
		return fmt.Errorf("%s next_loop %s is not in loop registry", c.ID, c.NextLoop)
	}
	if seen[c.ID] {
		return fmt.Errorf("duplicate chain id %s", c.ID)
	}
	seen[c.ID] = true
	if len(c.Changes) == 0 || len(c.Verifiers) == 0 || len(c.Evidence) == 0 {
		return fmt.Errorf("%s changes, verifiers, and evidence are required", c.ID)
	}
	for _, claim := range c.Claims {
		if !registry.Claims[claim] {
			return fmt.Errorf("%s claim %s is not in loop registry", c.ID, claim)
		}
		coveredClaims[claim] = true
	}
	if err := verifyRefs(root, c.ID, "changes", c.Changes); err != nil {
		return err
	}
	if err := verifyRefs(root, c.ID, "verifiers", c.Verifiers); err != nil {
		return err
	}
	if err := verifyRefs(root, c.ID, "evidence", c.Evidence); err != nil {
		return err
	}
	result.ChangeRefs += len(c.Changes)
	result.VerifierRefs += len(c.Verifiers)
	result.EvidenceRefs += len(c.Evidence)
	result.ClaimRefs += len(c.Claims)
	return nil
}

func verifyClaimCoverage(claims, covered map[string]bool) error {
	for claim := range claims {
		if !covered[claim] {
			return fmt.Errorf("loop registry claim %s has no evidence graph chain", claim)
		}
	}
	return nil
}
