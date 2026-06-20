package main

import "strings"

func workflowUploadArtifactSteps(text string) []workflowStep {
	lines := strings.Split(text, "\n")
	var steps []workflowStep
	for i := 0; i < len(lines); i++ {
		if !workflowLineUsesUploadArtifact(lines[i]) {
			continue
		}
		start := workflowStepStart(lines, i)
		end := workflowStepEnd(lines, start)
		steps = append(steps, workflowStep(lines[start:end]))
		i = end - 1
	}
	return steps
}

func workflowLineUsesUploadArtifact(line string) bool {
	return strings.Contains(workflowUsesValue(line), "actions/upload-artifact")
}

func workflowUsesValue(line string) string {
	trimmed := strings.TrimSpace(line)
	if value, ok := strings.CutPrefix(trimmed, "uses:"); ok {
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	if value, ok := strings.CutPrefix(trimmed, "- uses:"); ok {
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return ""
}
