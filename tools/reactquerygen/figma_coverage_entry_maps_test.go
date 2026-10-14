package main

func figmaCoverageEntriesByNode(entries []figmaSourceCoverageEntry) map[string]figmaSourceCoverageEntry {
	out := map[string]figmaSourceCoverageEntry{}
	for _, entry := range entries {
		out[entry.NodeID] = entry
	}
	return out
}
