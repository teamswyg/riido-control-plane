package main

import "fmt"

func verifyDecisionRules(m manifest) error {
	if len(m.DecisionRules) < 2 {
		return fmt.Errorf("keep and split decision rules are required")
	}
	actions := map[string]bool{}
	for _, rule := range m.DecisionRules {
		if rule.ThresholdDropPercent != minDecisionThreshold || rule.When == "" {
			return fmt.Errorf("invalid decision rule %+v", rule)
		}
		actions[rule.Action] = true
	}
	if !actions["keep_monolithic_snapshot"] || !actions["split_ai_agent_client_snapshot_only"] {
		return fmt.Errorf("decision rules must include keep and split actions")
	}
	return verifyCadenceAndSplit(m)
}

func verifyCadenceAndSplit(m manifest) error {
	if len(m.CadenceRules) < 2 {
		return fmt.Errorf("cadence rules are required")
	}
	for _, rule := range m.CadenceRules {
		if rule.Seconds <= 0 || rule.Seconds >= rule.MustStayBelowSeconds {
			return fmt.Errorf("cadence rule must stay below stale window: %+v", rule)
		}
	}
	if len(m.CandidateSplit.CommandModels) == 0 || len(m.CandidateSplit.QueryModels) == 0 {
		return fmt.Errorf("candidate split must name command and query models")
	}
	return nil
}

func verifyForbiddenAttributes(m manifest) error {
	for _, value := range []string{"task_id", "agent_id", "credentials", "payload_document"} {
		if !containsExact(m.ForbiddenTraceAttributes, value) {
			return fmt.Errorf("missing forbidden trace attribute %q", value)
		}
	}
	return nil
}
