package main

func newEvidence(m manifest) evidence {
	categoryCounts := map[string]int{}
	statusCounts := map[string]int{}
	partials := []partialCheck{}
	for _, check := range m.Checks {
		categoryCounts[check.Category]++
		statusCounts[check.Status]++
		if check.Status == "partial" {
			partials = append(partials, partialCheck{
				ID: check.ID, Category: check.Category,
				NextArtifact: check.NextArtifact, NextCommand: check.NextCommand,
			})
		}
	}
	return evidence{
		SchemaVersion: manifestSchemaToEvidence(m.SchemaVersion),
		Status:        "verified", CheckCount: len(m.Checks),
		CoveredCount: statusCounts["covered"], PartialCount: statusCounts["partial"],
		RequiredCategories: append([]string(nil), m.RequiredCategories...),
		MissingCategories:  missingCategories(m.RequiredCategories, categoryCounts),
		CategoryCounts:     categoryCounts, StatusCounts: statusCounts,
		PartialChecks: partials, Loop: m.Loop,
	}
}

func manifestSchemaToEvidence(schema string) string {
	if schema == manifestSchema {
		return evidenceSchema
	}
	return schema
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
