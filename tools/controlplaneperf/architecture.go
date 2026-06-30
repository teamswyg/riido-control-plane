package main

import (
	"fmt"
	"os"
)

var requiredPressureDimensions = []string{
	"heap_memory",
	"cpu_busy",
	"goroutine_delta",
	"race_condition",
	"otel_signal",
}

func verifyArchitectureComponents(root string, m manifest) error {
	if len(m.ArchitectureComponents) == 0 {
		return fmt.Errorf("performance manifest must declare architecture components")
	}
	seen := map[string]bool{}
	categories := architectureCategories(m.ArchitectureComponents)
	dimensions := architectureDimensions(m.ArchitectureComponents)
	for _, component := range m.ArchitectureComponents {
		if err := verifyArchitectureComponent(root, component, seen); err != nil {
			return err
		}
	}
	for _, path := range m.HotPaths {
		if !categories[path.Category] {
			return fmt.Errorf("hot path category %s lacks architecture component", path.Category)
		}
	}
	for _, dimension := range requiredPressureDimensions {
		if !dimensions[dimension] {
			return fmt.Errorf("pressure dimension %s lacks architecture component", dimension)
		}
	}
	return nil
}

func verifyArchitectureComponent(root string, c architectureComponent, seen map[string]bool) error {
	if c.ID == "" || c.Role == "" || seen[c.ID] {
		return fmt.Errorf("architecture component must have unique id and role")
	}
	seen[c.ID] = true
	if len(c.HotPathCategories) == 0 || len(c.Files) == 0 ||
		len(c.PressureDimensions) == 0 || len(c.ObservabilitySignals) == 0 ||
		len(c.EvidenceRefs) == 0 {
		return fmt.Errorf("architecture component %s missing semantic bindings", c.ID)
	}
	for _, file := range c.Files {
		if _, err := os.Stat(repoPath(root, file)); err != nil {
			return fmt.Errorf("architecture component %s missing file %s: %w", c.ID, file, err)
		}
	}
	return nil
}
