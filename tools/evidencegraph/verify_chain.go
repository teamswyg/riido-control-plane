package main

import "fmt"

func verifyChain(root string, c chain, seen map[string]bool, result *verifyResult, nextLoops map[string]bool) error {
	if c.ID == "" || c.Observation == "" || c.Hypothesis == "" {
		return fmt.Errorf("chain id, observation, and hypothesis are required")
	}
	if c.Decision == "" || c.NextLoop == "" {
		return fmt.Errorf("%s decision and next_loop are required", c.ID)
	}
	if !nextLoops[c.NextLoop] {
		return fmt.Errorf("%s next_loop %s is not in loop registry", c.ID, c.NextLoop)
	}
	if seen[c.ID] {
		return fmt.Errorf("duplicate chain id %s", c.ID)
	}
	seen[c.ID] = true
	if len(c.Changes) == 0 || len(c.Verifiers) == 0 || len(c.Evidence) == 0 {
		return fmt.Errorf("%s changes, verifiers, and evidence are required", c.ID)
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
	return nil
}
