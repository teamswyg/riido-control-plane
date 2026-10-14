package main

import "fmt"

func verifyCounts(want, got operationCounts) error {
	if got != want {
		return fmt.Errorf("operation counts = %+v, want %+v", got, want)
	}
	return nil
}

func verifyLoop(loop evidenceLoop) error {
	if loop.Observation == "" || loop.Hypothesis == "" || loop.Execute == "" ||
		loop.Evaluate == "" || loop.Retrospective == "" {
		return fmt.Errorf("evidence loop must define observe/hypothesis/execute/evaluate/retrospective")
	}
	return nil
}
