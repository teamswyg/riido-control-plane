package main

import "fmt"

func verifyPublicGates(repoRoot string, m manifest, result *verifyResult) error {
	seen := map[string]bool{}
	for _, gate := range m.PublicGates {
		if err := verifyPublicGate(repoRoot, m.ForbiddenDeps, gate, result); err != nil {
			return err
		}
		if seen[gate.Surface] {
			return fmt.Errorf("duplicate public gate %q", gate.Surface)
		}
		seen[gate.Surface] = true
	}
	return nil
}

func verifyPublicGate(repoRoot string, forbidden []string, gate publicGate, result *verifyResult) error {
	if gate.Surface == "" || gate.Verification == "" || gate.ExternalDependency == "" {
		return fmt.Errorf("public gate must define surface, verification, and external dependency")
	}
	if gate.PullRequestGate {
		result.PullRequestGates++
		if err := verifyNoForbiddenDependency(gate, forbidden); err != nil {
			return err
		}
	} else {
		result.OperatorGates++
	}
	result.CommandRefs += len(gate.Commands)
	for _, workflow := range gate.Workflows {
		result.WorkflowRefs++
		if err := verifyWorkflow(repoRoot, workflow, gate.PullRequestGate); err != nil {
			return fmt.Errorf("%s: %w", gate.Surface, err)
		}
	}
	return nil
}
