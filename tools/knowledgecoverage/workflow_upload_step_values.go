package main

import "strings"

func workflowStepNamesArtifact(step workflowStep, artifact string) bool {
	for _, line := range step {
		if workflowNameValue(line) == artifact {
			return true
		}
	}
	return false
}

func workflowStepFailsOnMissingFiles(step workflowStep) bool {
	for _, line := range step {
		if workflowIfNoFilesFoundValue(line) == "error" {
			return true
		}
	}
	return false
}

func workflowNameValue(line string) string {
	value, ok := strings.CutPrefix(strings.TrimSpace(line), "name:")
	if !ok {
		return ""
	}
	return strings.Trim(strings.TrimSpace(value), `"'`)
}
