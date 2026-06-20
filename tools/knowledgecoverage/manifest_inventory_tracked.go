package main

import "strings"

func trackedManifestSet(root string, m manifest, docs []docClass) map[string]bool {
	tracked := map[string]bool{}
	for _, doc := range docs {
		addTrackedManifest(root, tracked, siblingManifest(doc.Path))
	}
	for _, standalone := range m.Standalone {
		addTrackedManifest(root, tracked, standalone.Path)
	}
	for _, source := range m.SourceManifests {
		addTrackedManifest(root, tracked, source.Path)
	}
	for _, artifact := range m.ContractArtifacts {
		addTrackedManifest(root, tracked, artifact.Path)
	}
	for _, imported := range m.ImportedManifests {
		addTrackedManifest(root, tracked, imported.Path)
	}
	for _, owned := range m.OwnedManifests {
		addTrackedManifest(root, tracked, owned.Path)
	}
	return tracked
}

func addTrackedManifest(root string, tracked map[string]bool, path string) {
	if !strings.HasSuffix(path, ".riido.json") {
		return
	}
	if !fileExists(resolvePath(root, path)) {
		return
	}
	tracked[path] = true
}
