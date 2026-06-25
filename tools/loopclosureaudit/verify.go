package main

import (
	"fmt"
	"strings"
)

func verifyAll(root string, m manifest, deps dependencies) error {
	if m.SchemaVersion != manifestSchema || m.ID != "loop-closure-audit" {
		return fmt.Errorf("unexpected loop closure audit identity")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceTool != "tools/loopclosureaudit" {
		return fmt.Errorf("audit must bind doc, workflow, and evidence tool")
	}
	if m.EvidenceArtifact == "" || len(m.Requirements) == 0 {
		return fmt.Errorf("audit must declare evidence artifact and requirements")
	}
	if err := verifyLoopText(m.Loop); err != nil {
		return err
	}
	idx := newIndexes(deps)
	for _, req := range m.Requirements {
		if err := verifyRequirement(root, req, idx); err != nil {
			return err
		}
	}
	return verifyWorkflow(root, m)
}

func verifyLoopText(loop loopSpec) error {
	fields := []string{loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("audit loop must be complete")
		}
	}
	return nil
}

func verifyRequirement(root string, req requirement, idx indexes) error {
	if req.ID == "" || req.Statement == "" || len(req.Checks) == 0 {
		return fmt.Errorf("requirement must bind id, statement, and checks")
	}
	for _, c := range req.Checks {
		if err := verifyCheck(root, c, idx); err != nil {
			return fmt.Errorf("%s: %w", req.ID, err)
		}
	}
	return nil
}
