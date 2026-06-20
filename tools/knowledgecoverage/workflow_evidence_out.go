package main

import "strings"

func workflowEvidenceOutPaths(root, workflowPath, tool string) []string {
	data, err := readWorkflow(root, workflowPath)
	if err != nil {
		return nil
	}
	return workflowTextEvidenceOutPaths(string(data), tool)
}

func workflowTextEvidenceOutPaths(text, tool string) []string {
	var paths []string
	for _, command := range workflowRunCommands(text) {
		if workflowCommandRunsEvidenceTool(command, tool) {
			paths = append(paths, workflowCommandEvidenceOutPaths(command)...)
		}
	}
	return paths
}

func workflowCommandEvidenceOutPaths(command string) []string {
	fields := workflowCommandFields(command)
	var paths []string
	for i, field := range fields {
		if value, ok := strings.CutPrefix(field, "-evidence-out="); ok {
			paths = append(paths, workflowCleanCommandValue(value))
			continue
		}
		if field == "-evidence-out" && i+1 < len(fields) {
			paths = append(paths, workflowCleanCommandValue(fields[i+1]))
		}
	}
	return emptyIfNil(paths)
}

func workflowCommandFields(command string) []string {
	var fields []string
	for _, field := range strings.Fields(command) {
		if field != `\` {
			fields = append(fields, field)
		}
	}
	return fields
}

func workflowCleanCommandValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}
