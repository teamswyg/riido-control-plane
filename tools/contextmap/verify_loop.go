package main

import "fmt"

func verifyLoop(loop evidenceLoop) error {
	steps := map[string]string{
		"observation":   loop.Observation,
		"hypothesis":    loop.Hypothesis,
		"execute":       loop.Execute,
		"evaluate":      loop.Evaluate,
		"retrospective": loop.Retrospective,
	}
	for name, value := range steps {
		if value == "" {
			return fmt.Errorf("missing evidence loop step %q", name)
		}
	}
	return nil
}
