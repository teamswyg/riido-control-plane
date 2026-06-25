package main

import (
	"fmt"
	"strings"
)

func verifySource(root string, source promotionSource) error {
	if source.ID == "" || source.HarnessLoop == "" || source.PromotionTarget == "" {
		return fmt.Errorf("promotion source must bind id, harness loop, and target")
	}
	if source.SummaryPath == "" || source.CandidatePath == "" || source.CandidateArtifact == "" {
		return fmt.Errorf("promotion source %s must bind summary and candidate artifacts", source.ID)
	}
	if len(source.FailureStatuses) == 0 || len(source.RequiredNextArtifacts) == 0 {
		return fmt.Errorf("promotion source %s must define failure statuses and next artifacts", source.ID)
	}
	text, err := workflowText(root, source.SourceWorkflow)
	if err != nil {
		return err
	}
	if err := verifySourceWorkflow(text, source); err != nil {
		return err
	}
	return nil
}

func verifySourceWorkflow(text string, source promotionSource) error {
	required := []string{
		"./tools/harnesspromotion",
		"-summary " + source.SummaryPath,
		"-candidate-out " + source.CandidatePath,
		source.CandidatePath,
		"if-no-files-found: error",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("source %s workflow missing %q", source.ID, needle)
		}
	}
	if !workflowHasAlwaysStep(text, "./tools/harnesspromotion", "-summary "+source.SummaryPath, "-candidate-out "+source.CandidatePath) {
		return fmt.Errorf("source %s promotion step must run with if: always()", source.ID)
	}
	if !workflowHasAlwaysStep(text, "actions/upload-artifact", "name: "+source.CandidateArtifact, "path: "+source.CandidatePath, "if-no-files-found: error") {
		return fmt.Errorf("source %s candidate artifact upload must run with if: always()", source.ID)
	}
	return nil
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
