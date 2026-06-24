package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyLoopSchedule(root string, m manifest, loop loopRecord) error {
	if loop.ExpiresAfterHours > 24 {
		return nil
	}
	if strings.TrimSpace(loop.RefreshWorkflow) == "" {
		return fmt.Errorf("loop %s expires within 24h but has no refresh_workflow", loop.ID)
	}
	data, err := os.ReadFile(repoPath(root, loop.RefreshWorkflow))
	if err != nil {
		return fmt.Errorf("read refresh workflow %s: %w", loop.RefreshWorkflow, err)
	}
	text := string(data)
	if !strings.Contains(text, "schedule:") {
		return fmt.Errorf("loop %s expires within 24h but refresh workflow has no schedule", loop.ID)
	}
	if !refreshWorkflowPublishesEvidence(text, loop.Evidence) {
		return fmt.Errorf("loop %s refresh workflow must publish one strict loop evidence artifact", loop.ID)
	}
	return nil
}

func verifyEvidenceLoop(loop evidenceLoop) error {
	fields := []string{loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("evidence loop must be complete")
		}
	}
	return nil
}
