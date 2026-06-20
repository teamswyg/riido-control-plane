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
	for _, workflow := range workflows {
		if strings.Contains(workflow.Text, tool) &&
			strings.Contains(workflow.Text, "-check-doc") &&
			strings.Contains(workflow.Text, "-evidence-out") {
			return true
		}
	}
	return false
}
