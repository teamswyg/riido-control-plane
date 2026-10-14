package main

import "sort"

func pressureDimensionSummary(rows []architectureFileEvidence) []pressureDimensionEvidence {
	byDimension := map[string]*pressureDimensionEvidence{}
	for _, row := range rows {
		for _, dimension := range row.PressureDimensions {
			item := pressureDimensionRow(byDimension, dimension)
			item.Files = appendUnique(item.Files, row.Path)
			item.ComponentIDs = appendAllUnique(item.ComponentIDs, row.ComponentIDs)
			item.HotPathIDs = appendAllUnique(item.HotPathIDs, row.HotPathIDs)
			item.ObservabilitySignals = appendAllUnique(item.ObservabilitySignals, row.ObservabilitySignals)
			item.EvidenceRefs = appendAllUnique(item.EvidenceRefs, row.EvidenceRefs)
			item.TargetVerifierCommands = appendAllUnique(item.TargetVerifierCommands, row.TargetVerifierCommands)
		}
	}
	return sortedPressureDimensionRows(byDimension)
}

func pressureDimensionRow(
	byDimension map[string]*pressureDimensionEvidence,
	dimension string,
) *pressureDimensionEvidence {
	row := byDimension[dimension]
	if row == nil {
		row = &pressureDimensionEvidence{Dimension: dimension}
		byDimension[dimension] = row
	}
	return row
}

func sortedPressureDimensionRows(
	byDimension map[string]*pressureDimensionEvidence,
) []pressureDimensionEvidence {
	dimensions := make([]string, 0, len(byDimension))
	for dimension := range byDimension {
		dimensions = append(dimensions, dimension)
	}
	sort.Strings(dimensions)
	out := make([]pressureDimensionEvidence, 0, len(dimensions))
	for _, dimension := range dimensions {
		row := *byDimension[dimension]
		sort.Strings(row.ComponentIDs)
		sort.Strings(row.Files)
		sort.Strings(row.HotPathIDs)
		sort.Strings(row.ObservabilitySignals)
		sort.Strings(row.EvidenceRefs)
		sort.Strings(row.TargetVerifierCommands)
		out = append(out, row)
	}
	return out
}
