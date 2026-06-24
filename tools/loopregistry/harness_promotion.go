package main

import (
	"fmt"
	"os"
	"strings"
)

func captureHarnessPromotion(root string, loop loopRecord, result *verifyResult) error {
	if loop.Kind != kindHarness {
		return nil
	}
	data, err := os.ReadFile(repoPath(root, loop.RefreshWorkflow))
	if err != nil {
		return fmt.Errorf("read harness refresh workflow %s: %w", loop.RefreshWorkflow, err)
	}
	artifact, err := harnessCandidateArtifact(loop)
	if err != nil {
		return err
	}
	text := string(data)
	if !harnessWorkflowPromotesCandidates(text, artifact) {
		return fmt.Errorf("harness loop %s refresh workflow must run harnesspromotion and upload %s", loop.ID, artifact)
	}
	result.HarnessPromotionWorkflows[loop.ID] = loop.RefreshWorkflow
	result.HarnessCandidateArtifacts[loop.ID] = artifact
	return nil
}

func harnessWorkflowPromotesCandidates(text, artifact string) bool {
	return strings.Contains(text, "go run ./tools/harnesspromotion") &&
		strings.Contains(text, "-summary") &&
		strings.Contains(text, "-candidate-out") &&
		workflowUploadsStrictArtifact(text, artifact)
}
