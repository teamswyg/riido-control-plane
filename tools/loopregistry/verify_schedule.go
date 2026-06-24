package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyLoopSchedule(root string, m manifest, loop loopRecord) error {
	if strings.TrimSpace(loop.RefreshWorkflow) == "" {
		return fmt.Errorf("loop %s expires but has no refresh_workflow", loop.ID)
	}
	data, err := os.ReadFile(repoPath(root, loop.RefreshWorkflow))
	if err != nil {
		return fmt.Errorf("read refresh workflow %s: %w", loop.RefreshWorkflow, err)
	}
	text := string(data)
	if !strings.Contains(text, "schedule:") {
		return fmt.Errorf("loop %s expires but refresh workflow has no schedule", loop.ID)
	}
	cadence, err := refreshWorkflowCadenceMinutes(text)
	if err != nil {
		return fmt.Errorf("loop %s refresh workflow cadence: %w", loop.ID, err)
	}
	if cadence > expiryMinutes(loop) {
		return fmt.Errorf("loop %s refresh cadence %dm exceeds evidence expiry %dm",
			loop.ID, cadence, expiryMinutes(loop))
	}
	if !refreshWorkflowDeclaresLoopID(text, loop.ID) {
		return fmt.Errorf("loop %s refresh workflow must declare RIIDO_LOOP_IDS with its loop id", loop.ID)
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
