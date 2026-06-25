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
		"schedule:",
		"go test ./tools/loopclosureaudit",
		"go run ./tools/loopclosureaudit",
		"-check-doc",
		"-evidence-out out/loop-closure-audit-evidence.json",
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
