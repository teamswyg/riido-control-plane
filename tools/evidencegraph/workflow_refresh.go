package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func verifyEvidenceWorkflowRefresh(root string, m manifest, result *verifyResult) error {
	data, err := os.ReadFile(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read evidence workflow %s: %w", m.Workflow, err)
	}
	cadence, err := refreshCadenceMinutes(string(data))
	if err != nil {
		return fmt.Errorf("evidence workflow cadence: %w", err)
	}
	if cadence > evidenceGraphEvidenceTTLHours*60 {
		return fmt.Errorf("evidence workflow cadence %dm exceeds evidence ttl %dm",
			cadence, evidenceGraphEvidenceTTLHours*60)
	}
	result.RefreshCadenceMinutes = cadence
	if !workflowUploadsStrictArtifact(string(data), m.Evidence) {
		return fmt.Errorf("evidence workflow must upload strict artifact %s", m.Evidence)
	}
	return nil
}

func workflowFile(workflow string) string {
	return filepath.Base(workflow)
}

func manualRefreshCommand(workflow string) string {
	return "gh workflow run " + workflowFile(workflow) + " --ref main"
}
