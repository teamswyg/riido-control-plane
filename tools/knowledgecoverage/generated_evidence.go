package main

import (
	"fmt"
	"os"
)

func generatedToolCount(docs []docClass) int {
	count := 0
	for _, doc := range docs {
		if doc.Kind == "generated" && doc.GeneratorTool != "" {
			count++
		}
	}
	return count
}

func generatedEvidenceWorkflowCount(root string, docs []docClass) int {
	workflows := loadWorkflows(root)
	count := 0
	for _, doc := range docs {
		if doc.Kind == "generated" && doc.GeneratorTool != "" &&
			workflowHasGeneratorEvidence(workflows, doc.GeneratorTool) {
			count++
		}
	}
	return count
}

func generatedMissingTool(root string, docs []docClass) []string {
	var paths []string
	for _, doc := range docs {
		if doc.Kind != "generated" {
			continue
		}
		if doc.GeneratorTool == "" {
			paths = append(paths, doc.Path)
			continue
		}
		if _, err := os.Stat(resolvePath(root, doc.GeneratorTool)); err != nil {
			paths = append(paths, fmt.Sprintf("%s -> %s", doc.Path, doc.GeneratorTool))
		}
	}
	return emptyIfNil(paths)
}

func generatedMissingWorkflow(root string, docs []docClass) []string {
	workflows := loadWorkflows(root)
	var paths []string
	for _, doc := range docs {
		if doc.Kind == "generated" && doc.GeneratorTool != "" &&
			!workflowMentionsTool(workflows, doc.GeneratorTool) {
			paths = append(paths, fmt.Sprintf("%s -> %s", doc.Path, doc.GeneratorTool))
		}
	}
	return emptyIfNil(paths)
}

func generatedMissingEvidenceWorkflow(root string, docs []docClass) []string {
	workflows := loadWorkflows(root)
	var paths []string
	for _, doc := range docs {
		if doc.Kind == "generated" && doc.GeneratorTool != "" &&
			!workflowHasGeneratorEvidence(workflows, doc.GeneratorTool) {
			paths = append(paths, fmt.Sprintf("%s -> %s", doc.Path, doc.GeneratorTool))
		}
	}
	return emptyIfNil(paths)
}
