package main

import "sort"

func architectureFileIndex(
	components []architectureComponent,
	hotPaths []hotPath,
) []architectureFileEvidence {
	byPath := map[string]*architectureFileEvidence{}
	for _, component := range components {
		for _, path := range component.Files {
			row := architectureFileRow(byPath, path)
			row.ComponentIDs = appendUnique(row.ComponentIDs, component.ID)
			row.HotPathCategories = appendAllUnique(row.HotPathCategories, component.HotPathCategories)
			row.PressureDimensions = appendAllUnique(row.PressureDimensions, component.PressureDimensions)
			row.ObservabilitySignals = appendAllUnique(row.ObservabilitySignals, component.ObservabilitySignals)
			row.EvidenceRefs = appendAllUnique(row.EvidenceRefs, component.EvidenceRefs)
		}
	}
	applyHotPathRows(byPath, hotPaths)
	return sortedArchitectureFileRows(byPath)
}

func architectureFileRow(
	byPath map[string]*architectureFileEvidence,
	path string,
) *architectureFileEvidence {
	row := byPath[path]
	if row == nil {
		row = &architectureFileEvidence{Path: path}
		byPath[path] = row
	}
	return row
}

func appendAllUnique(dst, values []string) []string {
	for _, value := range values {
		dst = appendUnique(dst, value)
	}
	return dst
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func sortedArchitectureFileRows(byPath map[string]*architectureFileEvidence) []architectureFileEvidence {
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	out := make([]architectureFileEvidence, 0, len(paths))
	for _, path := range paths {
		row := *byPath[path]
		sort.Strings(row.ComponentIDs)
		sort.Strings(row.HotPathCategories)
		sort.Strings(row.PressureDimensions)
		sort.Strings(row.ObservabilitySignals)
		sort.Strings(row.EvidenceRefs)
		sort.Strings(row.HotPathIDs)
		sort.Strings(row.Benchmarks)
		sort.Strings(row.Tests)
		sort.Strings(row.OptimizationCandidates)
		sort.Strings(row.TargetVerifierCommands)
		out = append(out, row)
	}
	return out
}
