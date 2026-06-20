package main

import "fmt"

func validateStandaloneManifests(root string, m manifest) []string {
	var problems []string
	for _, standalone := range m.Standalone {
		problems = append(problems, validateStandaloneManifest(root, standalone)...)
	}
	return problems
}

func validateStandaloneManifest(root string, standalone standalone) []string {
	if standalone.Path == "" || standalone.EvidenceTool == "" ||
		standalone.Workflow == "" || standalone.EvidenceArtifact == "" {
		return []string{"standalone manifest path, evidence_tool, workflow, and evidence_artifact are required"}
	}
	if !standaloneManifestHasLoop(root, standalone.Path) {
		return []string{fmt.Sprintf("%s standalone manifest must define a complete evidence loop", standalone.Path)}
	}
	if !workflowRunsEvidenceTool(root, standalone.Workflow, standalone.EvidenceTool) {
		return []string{fmt.Sprintf("%s evidence_tool %q must run check-doc with evidence-out in %q",
			standalone.Path, standalone.EvidenceTool, standalone.Workflow)}
	}
	if !workflowUploadsEvidenceOutStrict(root, standalone.Workflow, standalone.EvidenceTool, standalone.EvidenceArtifact) {
		return []string{fmt.Sprintf("%s workflow %q must upload evidence_tool %q evidence-out path in strict artifact step %q",
			standalone.Path, standalone.Workflow, standalone.EvidenceTool, standalone.EvidenceArtifact)}
	}
	return nil
}

func standaloneManifestBindingCount(root string, m manifest) int {
	count := 0
	for _, standalone := range m.Standalone {
		if len(validateStandaloneManifest(root, standalone)) == 0 {
			count++
		}
	}
	return count
}

func standaloneManifestMissingBinding(root string, m manifest) []string {
	var paths []string
	for _, standalone := range m.Standalone {
		if len(validateStandaloneManifest(root, standalone)) > 0 {
			paths = append(paths, standalone.Path)
		}
	}
	return emptyIfNil(paths)
}
