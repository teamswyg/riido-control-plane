package main

func architectureCategories(components []architectureComponent) map[string]bool {
	out := map[string]bool{}
	for _, component := range components {
		for _, category := range component.HotPathCategories {
			out[category] = true
		}
	}
	return out
}

func architectureDimensions(components []architectureComponent) map[string]bool {
	out := map[string]bool{}
	for _, component := range components {
		for _, dimension := range component.PressureDimensions {
			out[dimension] = true
		}
	}
	return out
}

func architectureEvidenceRows(components []architectureComponent) []architectureComponentEvidence {
	rows := make([]architectureComponentEvidence, 0, len(components))
	for _, component := range components {
		rows = append(rows, architectureComponentEvidence{
			ID:                   component.ID,
			HotPathCategories:    component.HotPathCategories,
			PressureDimensions:   component.PressureDimensions,
			ObservabilitySignals: component.ObservabilitySignals,
			EvidenceRefs:         component.EvidenceRefs,
		})
	}
	return rows
}
