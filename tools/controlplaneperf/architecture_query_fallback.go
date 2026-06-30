package main

import (
	"path/filepath"
	"sort"
)

func architectureFallbackIndex(
	components []architectureComponent,
	hotPaths []hotPath,
) map[string]architectureFileEvidence {
	rows := architectureFileIndex(components, hotPaths)
	byDir := map[string]*architectureFileEvidence{}
	for _, row := range rows {
		dir := architecturePathDir(row.Path)
		target := architectureFallbackRow(byDir, dir)
		mergeArchitectureQueryFallback(target, row)
	}
	out := make(map[string]architectureFileEvidence, len(byDir))
	for dir, row := range byDir {
		sortArchitectureQueryFallback(row)
		out[dir] = *row
	}
	return out
}

func architecturePathDir(path string) string {
	return filepath.ToSlash(filepath.Dir(filepath.FromSlash(path)))
}

func architectureFallbackRow(
	byDir map[string]*architectureFileEvidence,
	dir string,
) *architectureFileEvidence {
	row := byDir[dir]
	if row == nil {
		row = &architectureFileEvidence{Path: dir}
		byDir[dir] = row
	}
	return row
}

func sortArchitectureQueryFallback(row *architectureFileEvidence) {
	sort.Strings(row.ComponentIDs)
	sort.Strings(row.HotPathIDs)
	sort.Strings(row.PressureDimensions)
	sort.Strings(row.ObservabilitySignals)
	sort.Strings(row.EvidenceRefs)
	sort.Strings(row.TargetVerifierCommands)
	sort.Strings(row.OptimizationCandidates)
}
