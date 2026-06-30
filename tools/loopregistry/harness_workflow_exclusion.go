package main

import (
	"fmt"
	"os"
	"strings"
)

func excludedHarnessWorkflows(root string, m manifest) (map[string]bool, error) {
	out := map[string]bool{}
	claims := claimsByID(m.Claims)
	for _, item := range m.HarnessWorkflowExclusions {
		workflow := strings.TrimSpace(item.Workflow)
		reason := strings.TrimSpace(item.Reason)
		replacementClaim := strings.TrimSpace(item.ReplacementClaim)
		replacementEvidence := strings.TrimSpace(item.ReplacementEvidence)
		if workflow == "" || reason == "" || replacementClaim == "" || replacementEvidence == "" {
			return nil, fmt.Errorf("harness workflow exclusions require workflow, reason, replacement_claim, and replacement_evidence")
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
		claim, ok := claims[replacementClaim]
		if !ok {
			return nil, fmt.Errorf("harness workflow exclusion %s replacement claim %s is missing", workflow, replacementClaim)
		}
		if !containsString(claim.Files, workflow) {
			return nil, fmt.Errorf("harness workflow exclusion %s replacement claim %s must bind workflow file", workflow, replacementClaim)
		}
		if !containsString(claim.CoversEvidence, replacementEvidence) {
			return nil, fmt.Errorf("harness workflow exclusion %s replacement claim %s must cover evidence %s", workflow, replacementClaim, replacementEvidence)
		}
		out[workflow] = true
	}
	return out, nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
