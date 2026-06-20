package main

import (
	"fmt"
	"os"
	"strings"
)

const domainDocPath = "docs/20-domain/saas-control-plane.md"

func verifyFocusedWorkflows(repo string, workflows []string) error {
	if len(workflows) < 10 {
		return fmt.Errorf("focused workflow coverage is underspecified: %d", len(workflows))
	}
	for _, workflow := range workflows {
		body, err := os.ReadFile(repoPath(repo, workflow))
		if err != nil {
			return fmt.Errorf("read focused workflow %q: %w", workflow, err)
		}
		text := string(body)
		if !strings.Contains(text, domainDocPath) {
			return fmt.Errorf("workflow %q does not watch %s", workflow, domainDocPath)
		}
		if !strings.Contains(text, "go run ./tools/dependencyallowlist") {
			return fmt.Errorf("workflow %q does not verify dependency allowlist", workflow)
		}
	}
	return nil
}

func verifyBoundaryWorkflow(m manifest, item boundary) error {
	if item.Workflow == "" {
		return fmt.Errorf("boundary %q missing workflow", item.ID)
	}
	if !stringSet(m.FocusedWorkflows)[item.Workflow] {
		return fmt.Errorf("boundary %q uses unregistered workflow %q", item.ID, item.Workflow)
	}
	return nil
}
