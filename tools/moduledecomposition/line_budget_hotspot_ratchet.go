package main

func buildLineBudgetHotspotRatchets(
	hotspots []lineBudgetHotspot,
	limits []lineBudgetHotspotLimit,
) []lineBudgetHotspotRatchet {
	byPath := lineBudgetHotspotIndex(hotspots)
	ratchets := make([]lineBudgetHotspotRatchet, 0, len(limits))
	for _, limit := range limits {
		current := byPath[limit.Path]
		ratchets = append(ratchets, lineBudgetHotspotRatchet{
			Path: limit.Path, Files: current.Files, MaxFiles: limit.MaxFiles,
			FilesSlack: limit.MaxFiles - current.Files,
			MaxLines:   current.MaxLines, MaxLinesLimit: limit.MaxLines,
			MaxLinesSlack: limit.MaxLines - current.MaxLines,
			TotalOver:     current.TotalOver, MaxTotalOver: limit.MaxTotalOver,
			TotalOverSlack: limit.MaxTotalOver - current.TotalOver,
		})
	}
	return ratchets
}

func lineBudgetHotspotIndex(hotspots []lineBudgetHotspot) map[string]lineBudgetHotspot {
	byPath := make(map[string]lineBudgetHotspot, len(hotspots))
	for _, hotspot := range hotspots {
		byPath[hotspot.Path] = hotspot
	}
	return byPath
}
