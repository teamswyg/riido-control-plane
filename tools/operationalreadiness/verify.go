package main

import (
	"fmt"
	"strings"
)

func verifyAll(root string, m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != "operational-readiness" {
		return fmt.Errorf("unexpected operational readiness manifest identity")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceTool != "tools/operationalreadiness" {
		return fmt.Errorf("readiness manifest must bind doc, workflow, and tool")
	}
	if m.EvidenceArtifact == "" || len(m.RequiredCategories) == 0 || len(m.Checks) == 0 {
		return fmt.Errorf("readiness manifest must bind artifact, categories, and checks")
	}
	if err := requireLocalFile(root, m.Workflow); err != nil {
		return err
	}
	if err := verifyChecks(root, m); err != nil {
		return err
	}
	return verifyLoop(m.Loop)
}

func verifyLoop(loop loopSpec) error {
	fields := []string{loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("readiness loop must be complete")
		}
	}
	return nil
}
