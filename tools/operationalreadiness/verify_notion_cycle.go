package main

import (
	"fmt"
	"strings"
)

func verifyNotionCycle(root string, checks map[string]bool, cycle notionCycle) error {
	if cycle.ID == "" || cycle.Priority == "" || cycle.Status == "" || cycle.CodexStatus == "" {
		return fmt.Errorf("notion cycle must bind identity and status")
	}
	if cycle.Priority != "P0" {
		return fmt.Errorf("notion cycle %s must preserve p0 priority", cycle.ID)
	}
	if cycle.Status != "partial" && cycle.Status != "covered" {
		return fmt.Errorf("notion cycle %s has unknown status %s", cycle.ID, cycle.Status)
	}
	if !strings.HasPrefix(cycle.Source, "notion:") || cycle.Summary == "" {
		return fmt.Errorf("notion cycle %s must bind source and summary", cycle.ID)
	}
	if !checks[cycle.BackfilledCheck] {
		return fmt.Errorf("notion cycle %s references unknown check %s", cycle.ID, cycle.BackfilledCheck)
	}
	if cycle.RequiredNextArtifact == "" || cycle.RequiredNextCommand == "" {
		return fmt.Errorf("notion cycle %s must bind next artifact and command", cycle.ID)
	}
	for _, ref := range cycle.EvidenceRefs {
		if err := verifyEvidenceRef(root, ref); err != nil {
			return fmt.Errorf("notion cycle %s: %w", cycle.ID, err)
		}
	}
	return nil
}
