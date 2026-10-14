package main

import (
	"fmt"
	"strings"
)

func verifyRenderedPhrases(m manifest, doc string, result *verifyResult) error {
	for _, phrase := range m.Required {
		result.PhraseChecks++
		if !strings.Contains(doc, phrase) {
			return fmt.Errorf("generated doc missing required phrase %q", phrase)
		}
	}
	for _, phrase := range m.Forbidden {
		result.ForbiddenChecks++
		if strings.Contains(doc, phrase) {
			return fmt.Errorf("generated doc contains forbidden stale phrase %q", phrase)
		}
	}
	return nil
}

func verifyLoop(loop loopRecord) error {
	if loop.Observation == "" || loop.Hypothesis == "" || loop.Execute == "" {
		return fmt.Errorf("loop observation, hypothesis, and execute are required")
	}
	if loop.Evaluate == "" || loop.Retrospective == "" {
		return fmt.Errorf("loop evaluate and retrospective are required")
	}
	return nil
}
