package main

import (
	"fmt"
	"strings"
)

func verifyWorkflow(root string, m manifest, result *verifyResult) error {
	text, err := readText(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read workflow: %w", err)
	}
	if !workflowScheduled(text) {
		return fmt.Errorf("workflow %s must run on a schedule", m.Workflow)
	}
	if !workflowUploadsEvidence(text, m.Evidence) {
		return fmt.Errorf("workflow %s must upload strict evidence artifact %s", m.Workflow, m.Evidence)
	}
	result.WorkflowScheduled = true
	return nil
}

func workflowScheduled(text string) bool {
	return strings.Contains(text, "schedule:") && strings.Contains(text, "cron:")
}

func workflowUploadsEvidence(text, artifact string) bool {
	return artifact != "" &&
		strings.Contains(text, "actions/upload-artifact") &&
		strings.Contains(text, "name: "+artifact) &&
		strings.Contains(text, "if-no-files-found: error")
}
