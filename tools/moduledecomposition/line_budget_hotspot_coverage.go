package main

func untrackedLineBudgetHotspots(
	hotspots []lineBudgetHotspot,
	limits []lineBudgetHotspotLimit,
) []lineBudgetHotspot {
	tracked := make(map[string]bool, len(limits))
	for _, limit := range limits {
		tracked[limit.Path] = true
	}
	out := make([]lineBudgetHotspot, 0)
	for _, hotspot := range hotspots {
		if !tracked[hotspot.Path] {
			out = append(out, hotspot)
		}
	}
	return out
}
