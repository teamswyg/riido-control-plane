package main

func categoryCounts(rows []surfaceEvidence) map[string]int {
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.Category]++
	}
	return counts
}

func missingCategories(required []string, counts map[string]int) []string {
	missing := []string{}
	for _, category := range required {
		if counts[category] == 0 {
			missing = append(missing, category)
		}
	}
	return missing
}
