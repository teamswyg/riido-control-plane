package main

import (
	"encoding/json"
	"os"
)

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
	path := resolvePath(root, generatedManifestPath(doc))
	data, err := os.ReadFile(path)
	if err != nil {
		return generatedManifestMeta{}, false
	}
	var meta generatedManifestMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return generatedManifestMeta{}, false
	}
	return meta, true
}
