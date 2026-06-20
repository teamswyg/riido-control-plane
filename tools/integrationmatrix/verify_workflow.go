package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyWorkflow(repoRoot, workflow string, pullRequestGate bool) error {
	if workflow == "" {
		return nil
	}
	path := repoPath(repoRoot, workflow)
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read workflow %s: %w", workflow, err)
	}
	hasPullRequest := strings.Contains(string(body), "pull_request:")
	if pullRequestGate && !hasPullRequest {
		return fmt.Errorf("%s must define pull_request trigger", workflow)
	}
	return nil
}

func verifyNoForbiddenDependency(gate publicGate, forbidden []string) error {
	haystack := strings.ToLower(gate.ExternalDependency + " " + gate.Verification)
	for _, phrase := range forbidden {
		needle := strings.ToLower(phrase)
		if strings.Contains(haystack, "no "+needle) {
			continue
		}
		if strings.Contains(haystack, strings.ToLower(phrase)) {
			return fmt.Errorf("%s pull_request gate has forbidden dependency %q", gate.Surface, phrase)
		}
	}
	return nil
}
