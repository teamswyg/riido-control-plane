package main

import "strings"

func workflowMentionsTool(workflows []workflowDoc, tool string) bool {
	for _, workflow := range workflows {
		if workflowTextRunsTool(workflow.Text, tool) {
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
	for _, command := range workflowRunCommands(text) {
		if workflowCommandRunsEvidenceTool(command, tool) {
			return true
		}
	}
	return false
}

func workflowTextRunsTool(text, tool string) bool {
	for _, command := range workflowRunCommands(text) {
		if strings.Contains(command, tool) {
			return true
		}
	}
	return false
}

func workflowCommandRunsEvidenceTool(command, tool string) bool {
	return strings.Contains(command, tool) &&
		strings.Contains(command, "-check-doc") &&
		strings.Contains(command, "-evidence-out")
}
