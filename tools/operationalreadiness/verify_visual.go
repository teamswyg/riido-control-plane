package main

import (
	"fmt"
	"strings"
)

func verifyVisualClosureEvidence(check readinessCheck) error {
	if check.Status != "covered" || !strings.Contains(check.NextArtifact, "visual_screenshot") {
		return nil
	}
	for _, measurement := range check.Measurements {
		if measurement.Kind == "screenshot" {
			return nil
		}
	}
	return fmt.Errorf("readiness check %s covered visual QA must bind screenshot evidence", check.ID)
}

func verifyVisualNextCommand(check readinessCheck) error {
	if !strings.Contains(check.NextArtifact, "visual_screenshot") {
		return nil
	}
	if strings.Contains(check.NextCommand, "body") && strings.Contains(check.NextCommand, "comments") {
		return nil
	}
	return fmt.Errorf("readiness check %s visual QA command must read issue body and comments", check.ID)
}
