package main

import "strings"

func workflowMentionsTool(workflows []workflowDoc, tool string) bool {
	for _, workflow := range workflows {
		if strings.Contains(workflow.Text, tool) {
			return true
		}
	}
	return false
}

func workflowHasGeneratorEvidence(workflows []workflowDoc, tool string) bool {
	return workflowHasEvidenceTool(workflows, tool)
}

func workflowHasEvidenceTool(workflows []workflowDoc, tool string) bool {
	for _, workflow := range workflows {
		if workflowTextRunsEvidenceTool(workflow.Text, tool) {
			return true
		}
	}
	return false
}

func workflowRunsEvidenceTool(root, workflowPath, tool string) bool {
	data, err := readWorkflow(root, workflowPath)
	if err != nil {
		return false
	}
	return workflowTextRunsEvidenceTool(string(data), tool)
}

func workflowTextRunsEvidenceTool(text, tool string) bool {
	return strings.Contains(text, tool) &&
		strings.Contains(text, "-check-doc") &&
		strings.Contains(text, "-evidence-out")
}
