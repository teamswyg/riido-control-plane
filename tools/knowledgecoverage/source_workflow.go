package main

import "strings"

func workflowRunsSourceEvidenceTool(root string, source sourceSSOT) bool {
	if pipeline, ok := readRiidoPipeline(root, source.Workflow); ok {
		for _, command := range riidoPipelineCommands(pipeline) {
			if workflowCommandRunsSourceEvidenceTool(command, source) {
				return true
			}
		}
		return false
	}
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
	if pipeline, ok := readRiidoPipeline(root, source.Workflow); ok {
		var paths []string
		for _, command := range riidoPipelineCommands(pipeline) {
			if workflowCommandRunsSourceEvidenceTool(command, source) {
				paths = append(paths, workflowCommandEvidenceOutPaths(command)...)
			}
		}
		return emptyIfNil(paths)
	}
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
	if pipeline, ok := readRiidoPipeline(root, source.Workflow); ok {
		for _, path := range workflowSourceEvidenceOutPaths(root, source) {
			if riidoPipelineUploadsStrict(pipeline, source.EvidenceArtifact, path) {
				return true
			}
		}
		return false
	}
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
