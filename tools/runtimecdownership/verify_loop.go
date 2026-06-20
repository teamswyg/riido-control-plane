package main

import "fmt"

func verifyLoop(loop evidenceLoop) (int, error) {
	if !nonEmpty(loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective) {
		return 0, fmt.Errorf("loop evidence must include observe, hypothesis, execute, evaluate, retrospective")
	}
	return 5, nil
}
