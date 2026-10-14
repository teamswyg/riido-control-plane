package main

import (
	"fmt"
	"strings"
)

func verifyWorkflowPathTriggers(text string, m manifest) error {
	for _, path := range workflowPathTriggers(m) {
		if path != "" && !strings.Contains(text, path) {
			return fmt.Errorf("candidate intake workflow missing semantic path trigger %q", path)
		}
	}
	return nil
}

func workflowPathTriggers(m manifest) []string {
	paths := []string{m.Workflow, m.GeneratedDoc}
	for _, source := range m.Sources {
		paths = append(paths,
			source.SourceWorkflow,
			source.ProducerManifest,
			source.LoopRegistryManifest,
			source.EvidenceGraphManifest,
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
