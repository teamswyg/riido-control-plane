package main

import "strings"

func refreshWorkflowPublishesEvidence(text string, evidence []evidenceSource) bool {
	for _, source := range evidence {
		if !source.Redacted {
			continue
		}
		if workflowUploadsStrictArtifact(text, source.Path) {
			return true
		}
	}
	return false
}

func workflowUploadsStrictArtifact(text, artifact string) bool {
	if artifact == "" || strings.Contains(artifact, "/") {
		return false
	}
	if !strings.Contains(text, "actions/upload-artifact") {
		return false
	}
	return strings.Contains(text, "name: "+artifact) &&
		strings.Contains(text, "if-no-files-found: error")
}
