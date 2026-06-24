package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyRegistryWorkflowCoversClaims(root string, m manifest) error {
	data, err := os.ReadFile(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read registry workflow: %w", err)
	}
	filters := workflowPathFilters(string(data))
	for _, path := range claimWorkflowPaths(m) {
		if !workflowPathCovered(filters, path) {
			return fmt.Errorf("loop registry workflow paths must include claim-bound path %s", path)
		}
	}
	return nil
}

func claimWorkflowPaths(m manifest) []string {
	seen := map[string]bool{}
	paths := []string{m.Workflow, m.GeneratedDoc, defaultManifest}
	for _, claim := range m.Claims {
		paths = append(paths, claim.Files...)
		paths = append(paths, claim.GeneratedDoc...)
	}
	out := []string{}
	for _, path := range paths {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	return out
}

func workflowPathCovered(filters map[string]bool, path string) bool {
	if filters[path] {
		return true
	}
	for filter := range filters {
		prefix, ok := strings.CutSuffix(filter, "/**")
		if ok && strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
