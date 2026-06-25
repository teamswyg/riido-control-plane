package main

import (
	"fmt"
	"strings"
)

func verifyWorkflow(root string, m manifest) error {
	text, err := readText(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read workflow %s: %w", m.Workflow, err)
	}
	required := []string{
		"go run ./tools/controlplaneaudit",
		"-evidence-out out/control-plane-high-traffic-audit.json",
		"name: " + m.EvidenceArtifact,
		"if-no-files-found: error",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("audit workflow missing %q", needle)
		}
	}
	return nil
}
