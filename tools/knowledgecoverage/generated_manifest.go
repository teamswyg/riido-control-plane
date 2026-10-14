package main

type generatedManifestMeta struct {
	GeneratedDoc     string `json:"generated_doc"`
	Workflow         string `json:"workflow"`
	EvidenceArtifact string `json:"evidence_artifact"`
	EvidenceTool     string `json:"evidence_tool"`
}

func generatedManifestPath(doc docClass) string {
	return siblingManifest(doc.Path)
}

func generatedManifestMetadata(root string, doc docClass) (generatedManifestMeta, bool) {
	return loadGeneratedManifestMeta(root, generatedManifestPath(doc))
}
