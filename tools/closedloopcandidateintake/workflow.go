package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyWorkflow(root string, m manifest) error {
	data, err := os.ReadFile(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read workflow %s: %w", m.Workflow, err)
	}
	text := string(data)
	required := []string{
		"schedule:",
		"go run ./tools/harnesspromotion",
		"go run ./tools/closedloopcandidateintake",
		"-candidate-in",
		"-evidence-out",
		"name: " + m.EvidenceArtifact,
		"if-no-files-found: error",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("candidate intake workflow missing %q", needle)
		}
	}
	return nil
}
