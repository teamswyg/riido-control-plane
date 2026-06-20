package main

func sourceManifestMetadataCount(root string, m manifest) int {
	count := 0
	for _, source := range m.SourceManifests {
		if sourceManifestHasMetadata(root, source.Path) {
			count++
		}
	}
	return count
}

func sourceManifestMissingMetadata(root string, m manifest) []string {
	var paths []string
	for _, source := range m.SourceManifests {
		if !sourceManifestHasMetadata(root, source.Path) {
			paths = append(paths, source.Path)
		}
	}
	return emptyIfNil(paths)
}
