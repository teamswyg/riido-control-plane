package main

import (
	"fmt"
	"os"
)

func validateDirectEvidence(root string, docs []docClass) []string {
	workflows := loadWorkflows(root)
	var problems []string
	for _, doc := range docs {
		if doc.Kind != "direct_ssot" {
			continue
		}
		if doc.EvidenceTool == "" {
			problems = append(problems, fmt.Sprintf("%s direct SSOT manifest must declare evidence_tool", doc.Path))
			continue
		}
		if _, err := os.Stat(resolvePath(root, doc.EvidenceTool)); err != nil {
			problems = append(problems, fmt.Sprintf("%s evidence_tool %q is missing", doc.Path, doc.EvidenceTool))
		}
		if !workflowHasEvidenceTool(workflows, doc.EvidenceTool) {
			problems = append(problems, fmt.Sprintf("%s evidence_tool %q must run check-doc with evidence-out in CI", doc.Path, doc.EvidenceTool))
		}
	}
	return problems
}

func directEvidenceWorkflowCount(root string, docs []docClass) int {
	workflows := loadWorkflows(root)
	count := 0
	for _, doc := range docs {
		if doc.Kind == "direct_ssot" && doc.EvidenceTool != "" &&
			workflowHasEvidenceTool(workflows, doc.EvidenceTool) {
			count++
		}
	}
	return count
}

func directMissingEvidenceWorkflow(root string, docs []docClass) []string {
	workflows := loadWorkflows(root)
	var paths []string
	for _, doc := range docs {
		if doc.Kind == "direct_ssot" && (doc.EvidenceTool == "" ||
			!workflowHasEvidenceTool(workflows, doc.EvidenceTool)) {
			paths = append(paths, fmt.Sprintf("%s -> %s", doc.Path, doc.EvidenceTool))
		}
	}
	return emptyIfNil(paths)
}
