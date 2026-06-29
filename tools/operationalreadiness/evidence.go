package main

import "time"

func newEvidence(m manifest) evidence {
	return newEvidenceAt(m, time.Now().UTC())
}

func newEvidenceAt(m manifest, now time.Time) evidence {
	categoryCounts := map[string]int{}
	measurementKinds := map[string]int{}
	statusCounts := map[string]int{}
	partials := []partialCheck{}
	measurementCount := 0
	stalePartials := 0
	for _, check := range m.Checks {
		categoryCounts[check.Category]++
		statusCounts[check.Status]++
		for _, measurement := range check.Measurements {
			measurementCount++
			measurementKinds[measurement.Kind]++
		}
		if check.Status == "partial" {
			ageDays := partialAgeDays(check.Date, now)
			stale := ageDays >= stalePartialAfterDays
			if stale {
				stalePartials++
			}
			partials = append(partials, partialCheck{
				ID: check.ID, Date: check.Date, Category: check.Category,
				AgeDays: ageDays, Stale: stale,
				NextArtifact: check.NextArtifact, NextCommand: check.NextCommand,
			})
		}
	}
	return evidence{
		SchemaVersion: manifestSchemaToEvidence(m.SchemaVersion),
		Status:        "verified", GeneratedAt: now.UTC().Format(time.RFC3339),
		CheckCount: len(m.Checks), MeasurementCount: measurementCount, CoveredCount: statusCounts["covered"],
		PartialCount: statusCounts["partial"], StalePartialCount: stalePartials,
		StaleAfterDays:     stalePartialAfterDays,
		RequiredCategories: append([]string(nil), m.RequiredCategories...),
		MissingCategories:  missingCategories(m.RequiredCategories, categoryCounts),
		CategoryCounts:     categoryCounts, MeasurementKinds: measurementKinds,
		StatusCounts:  statusCounts,
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
