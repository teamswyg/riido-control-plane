package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyEvidenceWorkflowCoversRefs(root string, m manifest) error {
	data, err := os.ReadFile(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read evidence workflow: %w", err)
	}
	filters := workflowPathFilters(string(data))
	for _, path := range evidenceWorkflowPaths(m) {
		if !workflowPathCovered(filters, path) {
			return fmt.Errorf("evidence graph workflow paths must include referenced path %s", path)
		}
	}
	return nil
}

func evidenceWorkflowPaths(m manifest) []string {
	seen := map[string]bool{}
	paths := []string{m.Workflow, m.GeneratedDoc, defaultManifest, m.LoopRegistry}
	for _, c := range m.Chains {
		paths = appendWorkflowRefPaths(paths, c.Changes)
		paths = appendWorkflowRefPaths(paths, c.Verifiers)
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

func appendWorkflowRefPaths(paths []string, refs []ref) []string {
	for _, item := range refs {
		if item.Kind != "artifact" {
			paths = append(paths, item.Path)
		}
	}
	return paths
}

func workflowPathCovered(filters map[string]bool, path string) bool {
	if filters[path] {
		return true
	}
	for filter := range filters {
		prefix, ok := strings.CutSuffix(filter, "/**")
		if ok && (path == prefix || strings.HasPrefix(path, prefix+"/")) {
			return true
		}
	}
	return false
}
