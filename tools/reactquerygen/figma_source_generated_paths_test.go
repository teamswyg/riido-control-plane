package main

func sourceCoverageGeneratedPathsByNode(manifest figmaSourceCoverageManifest) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, entry := range manifest.Entries {
		if _, ok := out[entry.NodeID]; !ok {
			out[entry.NodeID] = map[string]bool{}
		}
		for _, generatedPath := range entry.GeneratedPaths {
			out[entry.NodeID][generatedPath] = true
		}
	}
	return out
}
