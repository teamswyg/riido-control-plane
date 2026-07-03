package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/setutil"
)

const domainDocPath = "docs/20-domain/saas-control-plane.md"

func verifyFocusedWorkflows(repo string, workflows []string) error {
	if len(workflows) < 10 {
		return fmt.Errorf("focused workflow coverage is underspecified: %d", len(workflows))
	}
	for _, workflow := range workflows {
		body, err := os.ReadFile(pathutil.RepoPath(repo, workflow))
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

func verifyBoundaryWorkflow(repo string, m manifest, item boundary) error {
	if item.Workflow == "" {
		return fmt.Errorf("boundary %q missing workflow", item.ID)
	}
	if !setutil.StringSet(m.FocusedWorkflows)[item.Workflow] {
		return fmt.Errorf("boundary %q uses unregistered workflow %q", item.ID, item.Workflow)
	}
	if item.EvidenceArtifact != "" {
		return verifyBoundaryArtifact(repo, item)
	}
	return nil
}

func verifyBoundaryArtifact(repo string, item boundary) error {
	body, err := os.ReadFile(pathutil.RepoPath(repo, item.Workflow))
	if err != nil {
		return fmt.Errorf("read boundary workflow %q: %w", item.Workflow, err)
	}
	text := string(body)
	if !strings.Contains(text, item.EvidenceArtifact) || !strings.Contains(text, "evidence-out") {
		return fmt.Errorf("boundary %q workflow does not publish %s", item.ID, item.EvidenceArtifact)
	}
	return nil
}
