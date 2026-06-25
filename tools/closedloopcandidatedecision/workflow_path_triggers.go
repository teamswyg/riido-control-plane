package main

import (
	"fmt"
	"strings"
)

func verifyWorkflowPathTriggers(text string, m manifest, intake intakeManifest) error {
	for _, path := range workflowPathTriggers(m, intake) {
		if path != "" && !strings.Contains(text, path) {
			return fmt.Errorf("candidate decision workflow missing semantic path trigger %q", path)
		}
	}
	return nil
}

func workflowPathTriggers(m manifest, intake intakeManifest) []string {
	paths := []string{
		m.Workflow,
		m.GeneratedDoc,
		m.IntakeManifest,
		m.LoopRegistryManifest,
	}
	for _, source := range intake.Sources {
		paths = append(paths,
			source.ProducerManifest,
			source.LoopRegistryManifest,
		)
	}
	return uniqueStrings(paths)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
