package main

import (
	"fmt"
	"strings"
)

func verifyLoop(loop evidenceLoop) error {
	fields := []string{loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("evidence loop must be complete")
		}
	}
	return nil
}

func verifyDecisions(m manifest) error {
	if len(m.Decisions) == 0 || len(m.Assertions) == 0 {
		return fmt.Errorf("candidate decision must declare decisions and assertions")
	}
	seen := map[string]bool{}
	for _, decision := range m.Decisions {
		if seen[decision.CandidateID] {
			return fmt.Errorf("duplicate decision for candidate %s", decision.CandidateID)
		}
		seen[decision.CandidateID] = true
		if err := verifyDecision(decision); err != nil {
			return err
		}
	}
	return verifyDecisionTemplates(m.DecisionTemplates)
}
