package main

import "fmt"

func generatedMissingArtifactBinding(root string, docs []docClass) []string {
	var paths []string
	for _, doc := range docs {
		if doc.Kind != "generated" {
			continue
		}
		meta, ok := generatedManifestMetadata(root, doc)
		if !ok {
			paths = append(paths, fmt.Sprintf("%s -> %s", doc.Path, generatedManifestPath(doc)))
			continue
		}
		paths = append(paths, validateGeneratedManifestBinding(root, doc, meta)...)
	}
	return emptyIfNil(paths)
}
