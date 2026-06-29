package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyChecks(root string, m manifest) error {
	seen := map[string]bool{}
	categories := map[string]bool{}
	for _, check := range m.Checks {
		if err := verifyCheck(root, check); err != nil {
			return err
		}
		if seen[check.ID] {
			return fmt.Errorf("duplicate readiness check %s", check.ID)
		}
		seen[check.ID] = true
		categories[check.Category] = true
	}
	for _, category := range m.RequiredCategories {
		if !categories[category] {
			return fmt.Errorf("missing readiness category %s", category)
		}
	}
	return nil
}

func verifyCheck(root string, check readinessCheck) error {
	if check.ID == "" || check.Date == "" || check.Category == "" || check.Title == "" {
		return fmt.Errorf("readiness check must bind id, date, category, and title")
	}
	if _, err := readinessDate(check.Date); err != nil {
		return fmt.Errorf("readiness check %s has invalid date %q", check.ID, check.Date)
	}
	if check.Status != "covered" && check.Status != "partial" {
		return fmt.Errorf("readiness check %s has unknown status %s", check.ID, check.Status)
	}
	if len(check.EvidenceRefs) == 0 || check.NextArtifact == "" || check.NextCommand == "" {
		return fmt.Errorf("readiness check %s must bind evidence, next artifact, and command", check.ID)
	}
	if err := verifyMeasurements(check.ID, check.Measurements); err != nil {
		return err
	}
	for _, ref := range check.EvidenceRefs {
		if err := verifyEvidenceRef(root, ref); err != nil {
			return fmt.Errorf("readiness check %s: %w", check.ID, err)
		}
	}
	return nil
}

func verifyEvidenceRef(root string, ref evidenceRef) error {
	if ref.Kind == "" || ref.Path == "" {
		return fmt.Errorf("evidence ref must bind kind and path")
	}
	if strings.Contains(ref.Path, ":") {
		return nil
	}
	return requireLocalFile(root, ref.Path)
}

func requireLocalFile(root, path string) error {
	if _, err := os.Stat(repoPath(root, path)); err != nil {
		return fmt.Errorf("missing local evidence file %s: %w", path, err)
	}
	return nil
}
