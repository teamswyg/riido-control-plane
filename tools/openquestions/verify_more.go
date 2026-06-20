package main

import "fmt"

func verifyLoop(loop loopRecord) error {
	if loop.Observation == "" || loop.Hypothesis == "" || loop.Execute == "" {
		return fmt.Errorf("loop observation, hypothesis, and execute are required")
	}
	if loop.Evaluate == "" || loop.Retrospective == "" {
		return fmt.Errorf("loop evaluate and retrospective are required")
	}
	return nil
}
