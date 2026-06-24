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
	data, err := os.ReadFile(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read workflow %s: %w", m.Workflow, err)
	}
	if !strings.Contains(string(data), "schedule:") {
		return fmt.Errorf("loop %s expires within 24h but workflow has no schedule", loop.ID)
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
