package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func verifyHarnessWorkflowCoverage(root string, m manifest) error {
	workflows, err := harnessLikeWorkflowPaths(root)
	if err != nil {
		return err
	}
	registered := registeredHarnessWorkflows(m)
	excluded, err := excludedHarnessWorkflows(root, m)
	if err != nil {
		return err
	}
	for _, workflow := range workflows {
		if registered[workflow] || excluded[workflow] {
			continue
		}
		return fmt.Errorf("harness-like workflow %s must be registered as a harness loop or explicitly excluded", workflow)
	}
	return nil
}

func harnessLikeWorkflowPaths(root string) ([]string, error) {
	dir := repoPath(root, ".github/workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read workflow directory: %w", err)
	}
	out := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !harnessLikeWorkflowName(entry.Name()) {
			continue
		}
		out = append(out, ".github/workflows/"+entry.Name())
	}
	return out, nil
}

func harnessLikeWorkflowName(name string) bool {
	if filepath.Ext(name) != ".yml" {
		return false
	}
	base := strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name)))
	return strings.Contains(base, "smoke") ||
		strings.Contains(base, "load") ||
		strings.Contains(base, "harness")
}

func registeredHarnessWorkflows(m manifest) map[string]bool {
	out := map[string]bool{}
	for _, loop := range m.Loops {
		if loop.Kind == kindHarness && loop.RefreshWorkflow != "" {
			out[loop.RefreshWorkflow] = true
		}
	}
	return out
}
