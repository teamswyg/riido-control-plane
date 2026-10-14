package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyEvidenceRefreshWorkflows(root string, loop loopRecord) error {
	for _, source := range loop.Evidence {
		if !source.Redacted {
			continue
		}
		workflow := evidenceRefreshWorkflow(loop, source)
		data, err := os.ReadFile(repoPath(root, workflow))
		if err != nil {
			return fmt.Errorf("loop %s read evidence refresh workflow %s: %w", loop.ID, workflow, err)
		}
		text := string(data)
		if !strings.Contains(text, "schedule:") {
			return fmt.Errorf("loop %s evidence %s refresh workflow has no schedule", loop.ID, source.Path)
		}
		cadence, err := refreshWorkflowCadenceMinutes(text)
		if err != nil {
			return fmt.Errorf("loop %s evidence %s refresh workflow cadence: %w", loop.ID, source.Path, err)
		}
		if cadence > expiryMinutes(loop) {
			return fmt.Errorf("loop %s evidence %s refresh cadence %dm exceeds evidence expiry %dm",
				loop.ID, source.Path, cadence, expiryMinutes(loop))
		}
		if !refreshWorkflowDeclaresLoopID(text, loop.ID) {
			return fmt.Errorf("loop %s evidence %s refresh workflow must declare RIIDO_LOOP_IDS with its loop id",
				loop.ID, source.Path)
		}
		if !workflowUploadsStrictArtifact(text, source.Path) {
			return fmt.Errorf("loop %s evidence %s refresh workflow must upload strict artifact",
				loop.ID, source.Path)
		}
	}
	return nil
}

func workflowUploadsStrictArtifact(text, artifact string) bool {
	if artifact == "" || strings.Contains(artifact, "/") {
		return false
	}
	if !strings.Contains(text, "actions/upload-artifact") {
		return false
	}
	return strings.Contains(text, "name: "+artifact) &&
		strings.Contains(text, "if-no-files-found: error")
}
