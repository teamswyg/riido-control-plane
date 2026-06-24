package main

import (
	"fmt"
	"os"
	"strings"
)

func excludedHarnessWorkflows(root string, m manifest) (map[string]bool, error) {
	out := map[string]bool{}
	for _, item := range m.HarnessWorkflowExclusions {
		workflow := strings.TrimSpace(item.Workflow)
		if workflow == "" || strings.TrimSpace(item.Reason) == "" {
			return nil, fmt.Errorf("harness workflow exclusions require workflow and reason")
		}
		if !harnessLikeWorkflowName(workflow) {
			return nil, fmt.Errorf("harness workflow exclusion %s is not harness-like", workflow)
		}
		if _, err := os.Stat(repoPath(root, workflow)); err != nil {
			return nil, fmt.Errorf("harness workflow exclusion %s is stale: %w", workflow, err)
		}
		if out[workflow] {
			return nil, fmt.Errorf("harness workflow exclusion %s is duplicated", workflow)
		}
		out[workflow] = true
	}
	return out, nil
}
