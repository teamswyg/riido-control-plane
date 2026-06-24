package main

import (
	"fmt"
	"strings"
)

func verifySource(root string, source intakeSource) error {
	if source.ID == "" || source.CandidateArtifact == "" || source.PromotionTarget == "" {
		return fmt.Errorf("source must bind id, candidate artifact, and promotion target")
	}
	if err := verifyRequiredNextArtifacts(source.RequiredNextArtifacts, source.ID); err != nil {
		return err
	}
	if err := verifyProducer(root, source); err != nil {
		return err
	}
	if err := verifyLoopRegistry(root, source); err != nil {
		return err
	}
	return verifyEvidenceGraph(root, source)
}

func verifyLoop(loop evidenceLoop) error {
	fields := []string{loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("evidence loop must be complete")
		}
	}
	return nil
}
