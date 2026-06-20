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
	if meta.Workflow == "" || meta.EvidenceArtifact == "" {
		return []string{fmt.Sprintf("%s manifest must declare workflow and evidence_artifact", doc.Path)}
	}
	if !workflowUploadsArtifact(root, meta.Workflow, meta.EvidenceArtifact) {
		return []string{fmt.Sprintf("%s workflow %q must upload evidence artifact %q",
			doc.Path, meta.Workflow, meta.EvidenceArtifact)}
	}
	if !workflowUploadsArtifactStrict(root, meta.Workflow, meta.EvidenceArtifact) {
		return []string{fmt.Sprintf("%s workflow %q must fail when evidence artifact %q is missing",
			doc.Path, meta.Workflow, meta.EvidenceArtifact)}
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
