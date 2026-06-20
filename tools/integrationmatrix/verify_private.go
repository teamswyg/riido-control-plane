package main

import "fmt"

func verifyPrivateGates(gates []privateGate) error {
	seen := map[string]bool{}
	for _, gate := range gates {
		if gate.Surface == "" || gate.Owner == "" || gate.Evidence == "" {
			return fmt.Errorf("private gate must define surface, owner, and evidence")
		}
		if seen[gate.Surface] {
			return fmt.Errorf("duplicate private gate %q", gate.Surface)
		}
		seen[gate.Surface] = true
	}
	return nil
}

func verifyLoop(loop evidenceLoop) error {
	if loop.Observation == "" || loop.Hypothesis == "" || loop.Execute == "" ||
		loop.Evaluate == "" || loop.Retrospective == "" {
		return fmt.Errorf("loop must define observe/hypothesis/execute/evaluate/retrospective")
	}
	return nil
}
