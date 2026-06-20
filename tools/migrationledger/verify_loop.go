package main

import (
	"fmt"
	"strings"
)

func verifyLoop(loop evidenceLoop) error {
	values := map[string]string{
		"observation":   loop.Observation,
		"hypothesis":    loop.Hypothesis,
		"execute":       loop.Execute,
		"evaluate":      loop.Evaluate,
		"retrospective": loop.Retrospective,
	}
	for name, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("loop.%s is required", name)
		}
	}
	return nil
}
