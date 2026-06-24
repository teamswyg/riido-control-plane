package main

import "fmt"

func verifyClaimSpecs(spec workflowSpec) error {
	seen := map[string]bool{}
	for _, claim := range spec.EvidenceClaims {
		if claim.ID == "" || claim.Summary == "" {
			return fmt.Errorf("workflow %q has incomplete evidence claim", spec.ID)
		}
		if seen[claim.ID] {
			return fmt.Errorf("workflow %q has duplicate evidence claim %q", spec.ID, claim.ID)
		}
		seen[claim.ID] = true
		if len(claim.SourcePhrases) == 0 {
			return fmt.Errorf("workflow %q claim %q needs source phrases", spec.ID, claim.ID)
		}
	}
	return nil
}

func claimSourcePhrases(spec workflowSpec) []string {
	phrases := []string{}
	for _, claim := range spec.EvidenceClaims {
		phrases = append(phrases, claim.SourcePhrases...)
	}
	return phrases
}
