package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyOperationalWorkflow(root string, m manifest) error {
	data, err := os.ReadFile(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read workflow %s: %w", m.Workflow, err)
	}
	text := string(data)
	required := []string{
		"-candidate-out out/operational-readiness-closed-loop-candidates.json",
		"name: operational-readiness-closed-loop-candidates",
		"out/operational-readiness-closed-loop-candidates.json",
		"if-no-files-found: error",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("operational readiness workflow missing %q", needle)
		}
	}
	return nil
}
