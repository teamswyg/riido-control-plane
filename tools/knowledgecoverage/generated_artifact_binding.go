package main

import "fmt"

func validateGeneratedArtifactBindings(root string, docs []docClass) []string {
	var problems []string
	for _, doc := range docs {
		if doc.Kind != "generated" {
			continue
		}
		meta, ok := generatedManifestMetadata(root, doc)
		if !ok {
			problems = append(problems, fmt.Sprintf("generated doc %q must have reader manifest %q",
				doc.Path, generatedManifestPath(doc)))
			continue
		}
		problems = append(problems, validateGeneratedManifestBinding(root, doc, meta)...)
	}
	return problems
}

func validateGeneratedManifestBinding(root string, doc docClass, meta generatedManifestMeta) []string {
	if meta.GeneratedDoc != doc.Path {
		return []string{fmt.Sprintf("%s manifest generated_doc mismatch: %q", doc.Path, meta.GeneratedDoc)}
	}
	if meta.EvidenceTool != "" && meta.EvidenceTool != doc.GeneratorTool {
		return []string{fmt.Sprintf("%s manifest evidence_tool %q must match generator %q",
			doc.Path, meta.EvidenceTool, doc.GeneratorTool)}
	}
	if meta.Workflow == "" || meta.EvidenceArtifact == "" {
		return []string{fmt.Sprintf("%s manifest must declare workflow and evidence_artifact", doc.Path)}
	}
	if !workflowRunsEvidenceTool(root, meta.Workflow, doc.GeneratorTool) {
		return []string{fmt.Sprintf("%s workflow %q must run generator %q with check-doc and evidence-out",
			doc.Path, meta.Workflow, doc.GeneratorTool)}
	}
	if !workflowUploadsArtifact(root, meta.Workflow, meta.EvidenceArtifact) {
		return []string{fmt.Sprintf("%s workflow %q must upload evidence artifact %q",
			doc.Path, meta.Workflow, meta.EvidenceArtifact)}
	}
	if !workflowUploadsArtifactStrict(root, meta.Workflow, meta.EvidenceArtifact) {
		return []string{fmt.Sprintf("%s workflow %q must fail when evidence artifact %q is missing",
			doc.Path, meta.Workflow, meta.EvidenceArtifact)}
	}
	if !workflowUploadsEvidenceOut(root, meta.Workflow, doc.GeneratorTool, meta.EvidenceArtifact) {
		return []string{fmt.Sprintf("%s workflow %q must upload generator %q evidence-out path in artifact %q",
			doc.Path, meta.Workflow, doc.GeneratorTool, meta.EvidenceArtifact)}
	}
	if !workflowUploadsEvidenceOutStrict(root, meta.Workflow, doc.GeneratorTool, meta.EvidenceArtifact) {
		return []string{fmt.Sprintf("%s workflow %q must upload generator %q evidence-out path in strict artifact step %q",
			doc.Path, meta.Workflow, doc.GeneratorTool, meta.EvidenceArtifact)}
	}
	return nil
}

func generatedArtifactBindingCount(root string, docs []docClass) int {
	count := 0
	for _, doc := range docs {
		if doc.Kind != "generated" {
			continue
		}
		meta, ok := generatedManifestMetadata(root, doc)
		if ok && len(validateGeneratedManifestBinding(root, doc, meta)) == 0 {
			count++
		}
	}
	return count
}
