package main

import (
	"fmt"
	"strings"
)

var (
	allowedDispositions = []string{"adopted", "triage_required", "deferred", "rejected"}
	allowedPriorities   = []string{"P0", "P1", "P2", "P3"}
)

func verifyDecision(decision decisionRecord) error {
	required := []string{
		decision.CandidateID, decision.Disposition, decision.Priority,
		decision.Owner, decision.NextLoop, decision.NextArtifact, decision.Reason,
	}
	for _, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("candidate decision fields must be complete")
		}
	}
	if !containsString(allowedDispositions, decision.Disposition) {
		return fmt.Errorf("candidate %s has unknown disposition %s", decision.CandidateID, decision.Disposition)
	}
	if !containsString(allowedPriorities, decision.Priority) {
		return fmt.Errorf("candidate %s has unknown priority %s", decision.CandidateID, decision.Priority)
	}
	return nil
}

func containsString(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
