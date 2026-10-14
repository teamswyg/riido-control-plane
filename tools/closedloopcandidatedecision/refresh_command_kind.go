package main

import "strings"

func refreshCommandKind(command string) string {
	if strings.HasPrefix(strings.TrimSpace(command), "gh workflow run ") {
		return "refresh_workflow"
	}
	return "target_verifier"
}

func refreshLoopID(item decisionArtifactEvidence) string {
	if item.NextLoop != "" {
		return item.NextLoop
	}
	return "closed_loop_candidate_decision"
}
