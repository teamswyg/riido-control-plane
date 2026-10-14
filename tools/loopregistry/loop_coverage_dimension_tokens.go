package main

func loopEvidenceSourcePaths(loop loopRecord) []string {
	paths := make([]string, 0, len(loop.Evidence))
	for _, source := range loop.Evidence {
		paths = append(paths, source.Path)
	}
	return paths
}

func loopCoverageTokens(
	loops []loopRecord,
	dim loopCoverageDimension,
) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, loop := range loops {
		out[loop.ID] = map[string]bool{}
		for _, token := range dim.loopTokens(loop) {
			out[loop.ID][token] = true
		}
	}
	return out
}
