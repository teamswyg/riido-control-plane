package main

import (
	"fmt"
	"strings"
)

func verifyWorkflow(root string, m manifest) error {
	text, err := readText(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read workflow %s: %w", m.Workflow, err)
	}
	required := []string{
		"schedule:",
		"go run ./tools/controlplaneperf",
		"go run ./tools/controlplanepressure",
		"go test ./internal/riidoaiserver",
		"-benchmem",
		"benchtime=100ms",
		"go run ./tools/liveworkflowevidence",
		"go run ./tools/harnesspromotion",
		"name: " + m.EvidenceArtifact,
		"name: " + m.BenchmarkArtifact,
		"name: " + m.LocalPressureArtifact,
		"name: " + m.SummaryArtifact,
		"name: " + m.CandidateArtifact,
		"if-no-files-found: error",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("performance workflow missing %q", needle)
		}
	}
	return nil
}

func verifyLoop(loop loopSpec) error {
	fields := []string{loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("performance loop must be complete")
		}
	}
	return nil
}
