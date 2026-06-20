package main

func compareLineBudgetHotspots(a, b lineBudgetHotspot) int {
	if a.Files != b.Files {
		return b.Files - a.Files
	}
	if a.TotalOver != b.TotalOver {
		return b.TotalOver - a.TotalOver
	}
	if a.MaxLines != b.MaxLines {
		return b.MaxLines - a.MaxLines
	}
	if a.Path < b.Path {
		return -1
	}
	if a.Path > b.Path {
		return 1
	}
	return 0
}

func trimLineBudgetHotspots(hotspots []lineBudgetHotspot, limit int) []lineBudgetHotspot {
	if limit > 0 && len(hotspots) > limit {
		return hotspots[:limit]
	}
	return hotspots
}
