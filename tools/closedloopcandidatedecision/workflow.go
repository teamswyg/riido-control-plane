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
		"actions: read",
		"schedule:",
		"gh run download",
		"live-candidate-files.txt",
		"RIIDO_HARNESS_PROMOTION_NOW=2026-06-24T12:00:00Z",
		"go run ./tools/harnesspromotion",
		"go run ./tools/closedloopcandidateintake",
		"go run ./tools/closedloopcandidatedecision",
		"-candidate-in",
		"name: " + m.EvidenceArtifact,
		"if-no-files-found: error",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("candidate decision workflow missing %q", needle)
		}
	}
	var intake intakeManifest
	if err := readJSON(repoPath(root, m.IntakeManifest), &intake); err != nil {
		return err
	}
	for _, source := range intake.Sources {
		if source.CandidateArtifact == "" || !strings.Contains(text, source.CandidateArtifact) {
			return fmt.Errorf("candidate decision workflow missing live artifact for source %q", source.ID)
		}
	}
	return verifyWorkflowPathTriggers(text, m, intake)
}
