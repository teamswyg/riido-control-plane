package main

import "strings"

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
