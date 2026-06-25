package main

import (
	"fmt"
	"strings"
)

func verifySurfaces(root string, surfaces []surface) error {
	seen := map[string]bool{}
	for _, surface := range surfaces {
		if surface.ID == "" || surface.Category == "" || surface.Risk == "" || surface.Candidate == "" {
			return fmt.Errorf("surface must bind id, category, risk, and candidate")
		}
		if seen[surface.ID] {
			return fmt.Errorf("duplicate surface %s", surface.ID)
		}
		seen[surface.ID] = true
		if err := verifySurfaceFiles(root, surface); err != nil {
			return err
		}
	}
	return nil
}

func verifySurfaceFiles(root string, surface surface) error {
	if len(surface.Files) == 0 || len(surface.Patterns) == 0 {
		return fmt.Errorf("surface %s must bind files and patterns", surface.ID)
	}
	text := strings.Builder{}
	for _, file := range surface.Files {
		if !fileExists(repoPath(root, file)) {
			return fmt.Errorf("surface %s missing file %s", surface.ID, file)
		}
		data, err := readText(repoPath(root, file))
		if err != nil {
			return err
		}
		text.WriteString(data)
	}
	return verifyPatterns(surface, text.String())
}

func verifyPatterns(surface surface, text string) error {
	for _, pattern := range surface.Patterns {
		if !strings.Contains(text, pattern) {
			return fmt.Errorf("surface %s missing pattern %q", surface.ID, pattern)
		}
	}
	return nil
}
