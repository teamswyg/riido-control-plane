package main

import "strings"

func workflowRunsSourceEvidenceTool(root string, source sourceSSOT) bool {
	data, err := readWorkflow(root, source.Workflow)
	if err != nil {
		return false
	}
	for _, command := range workflowRunCommands(string(data)) {
		if workflowCommandRunsSourceEvidenceTool(command, source) {
			return true
		}
	}
	return false
}

func workflowCommandRunsSourceEvidenceTool(command string, source sourceSSOT) bool {
	return strings.Contains(command, source.EvidenceTool) &&
		strings.Contains(command, source.Path) &&
		strings.Contains(command, "-evidence-out")
}

func workflowSourceEvidenceOutPaths(root string, source sourceSSOT) []string {
	data, err := readWorkflow(root, source.Workflow)
	if err != nil {
		return nil
	}
	var paths []string
	for _, command := range workflowRunCommands(string(data)) {
		if workflowCommandRunsSourceEvidenceTool(command, source) {
			paths = append(paths, workflowCommandEvidenceOutPaths(command)...)
		}
	}
	return emptyIfNil(paths)
}

func workflowUploadsSourceEvidenceOutStrict(root string, source sourceSSOT) bool {
	data, err := readWorkflow(root, source.Workflow)
	if err != nil {
		return false
	}
	for _, path := range workflowSourceEvidenceOutPaths(root, source) {
		if workflowTextUploadsArtifactPathStrict(string(data), source.EvidenceArtifact, path) {
			return true
		}
	}
	return false
}
