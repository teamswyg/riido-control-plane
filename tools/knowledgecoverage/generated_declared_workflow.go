package main

import "fmt"

func generatedDeclaredWorkflowEvidenceCount(root string, docs []docClass) int {
	count := 0
	for _, doc := range docs {
		if doc.Kind != "generated" || doc.GeneratorTool == "" {
			continue
		}
		meta, ok := generatedManifestMetadata(root, doc)
		if ok && workflowRunsEvidenceTool(root, meta.Workflow, doc.GeneratorTool) {
			count++
		}
	}
	return count
}

func generatedMissingDeclaredWorkflowEvidence(root string, docs []docClass) []string {
	var paths []string
	for _, doc := range docs {
		if doc.Kind != "generated" || doc.GeneratorTool == "" {
			continue
		}
		meta, ok := generatedManifestMetadata(root, doc)
		if !ok || workflowRunsEvidenceTool(root, meta.Workflow, doc.GeneratorTool) {
			continue
		}
		paths = append(paths, fmt.Sprintf("%s -> %s -> %s", doc.Path, meta.Workflow, doc.GeneratorTool))
	}
	return emptyIfNil(paths)
}
