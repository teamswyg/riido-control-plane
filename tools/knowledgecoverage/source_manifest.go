package main

import "fmt"

func validateSourceManifests(root string, m manifest) []string {
	var problems []string
	for _, source := range m.SourceManifests {
		problems = append(problems, validateSourceManifest(root, source)...)
	}
	return problems
}

func validateSourceManifest(root string, source sourceSSOT) []string {
	if source.Path == "" || source.EvidenceTool == "" ||
		source.Workflow == "" || source.EvidenceArtifact == "" {
		return []string{"source manifest path, evidence_tool, workflow, and evidence_artifact are required"}
	}
	if !sourceManifestHasMetadata(root, source.Path) {
		return []string{fmt.Sprintf("%s source manifest must define id, assertions, and a complete evidence loop",
			source.Path)}
	}
	if !workflowRunsSourceEvidenceTool(root, source) {
		return []string{fmt.Sprintf("%s evidence_tool %q must read source path and write evidence-out in %q",
			source.Path, source.EvidenceTool, source.Workflow)}
	}
	if !workflowUploadsSourceEvidenceOutStrict(root, source) {
		return []string{fmt.Sprintf("%s workflow %q must upload evidence_tool %q evidence-out path in strict artifact step %q",
			source.Path, source.Workflow, source.EvidenceTool, source.EvidenceArtifact)}
	}
	return nil
}

func sourceManifestBindingCount(root string, m manifest) int {
	count := 0
	for _, source := range m.SourceManifests {
		if len(validateSourceManifest(root, source)) == 0 {
			count++
		}
	}
	return count
}

func sourceManifestMissingBinding(root string, m manifest) []string {
	var paths []string
	for _, source := range m.SourceManifests {
		if len(validateSourceManifest(root, source)) > 0 {
			paths = append(paths, source.Path)
		}
	}
	return emptyIfNil(paths)
}
