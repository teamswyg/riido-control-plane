package main

import "fmt"

func generatedManifestEvidenceToolCount(root string, docs []docClass) int {
	count := 0
	for _, doc := range docs {
		if doc.Kind != "generated" {
			continue
		}
		meta, ok := generatedManifestMetadata(root, doc)
		if ok && meta.EvidenceTool != "" {
			count++
		}
	}
	return count
}

func generatedManifestEvidenceToolMismatch(root string, docs []docClass) []string {
	var paths []string
	for _, doc := range docs {
		if doc.Kind != "generated" {
			continue
		}
		meta, ok := generatedManifestMetadata(root, doc)
		if ok && meta.EvidenceTool != "" && meta.EvidenceTool != doc.GeneratorTool {
			paths = append(paths, fmt.Sprintf("%s -> %s != %s", doc.Path, meta.EvidenceTool, doc.GeneratorTool))
		}
	}
	return emptyIfNil(paths)
}
