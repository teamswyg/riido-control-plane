package main

import (
	"errors"
	"fmt"
	"strings"
)

func verifyEvidenceMetadata(id string, assertions []string, loop evidenceLoop) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}
	if len(assertions) == 0 {
		return errors.New("assertions are required")
	}
	for i, assertion := range assertions {
		if strings.TrimSpace(assertion) == "" {
			return fmt.Errorf("assertions[%d] is required", i)
		}
	}
	for name, value := range map[string]string{
		"observation":   loop.Observation,
		"hypothesis":    loop.Hypothesis,
		"execute":       loop.Execute,
		"evaluate":      loop.Evaluate,
		"retrospective": loop.Retrospective,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("loop.%s is required", name)
		}
	}
	return nil
}
