package main

const architectureQuerySchema = "riido-control-plane-performance-architecture-query.v1"

func newArchitectureQuery(m manifest, paths []string) architectureQueryEvidence {
	index := architectureRowsByPath(architectureFileIndex(m.ArchitectureComponents, m.HotPaths))
	rows := make([]architectureQueryRow, 0, len(paths))
	hits := 0
	for _, path := range paths {
		row := architectureQueryRow{Path: path}
		if indexed, ok := index[path]; ok {
			row = architectureQueryRow{
				Path:                   path,
				Matched:                true,
				ComponentIDs:           append([]string(nil), indexed.ComponentIDs...),
				HotPathIDs:             append([]string(nil), indexed.HotPathIDs...),
				PressureDimensions:     append([]string(nil), indexed.PressureDimensions...),
				ObservabilitySignals:   append([]string(nil), indexed.ObservabilitySignals...),
				TargetVerifierCommands: append([]string(nil), indexed.TargetVerifierCommands...),
				OptimizationCandidates: append([]string(nil), indexed.OptimizationCandidates...),
			}
			hits++
		}
		rows = append(rows, row)
	}
	return architectureQueryEvidence{
		SchemaVersion: architectureQuerySchema,
		Status:        queryStatus(len(paths), hits),
		QueryCount:    len(paths),
		HitCount:      hits,
		MissCount:     len(paths) - hits,
		Queries:       rows,
	}
}

func architectureRowsByPath(rows []architectureFileEvidence) map[string]architectureFileEvidence {
	out := make(map[string]architectureFileEvidence, len(rows))
	for _, row := range rows {
		out[row.Path] = row
	}
	return out
}

func queryStatus(total, hits int) string {
	if total == hits {
		return "matched"
	}
	if hits == 0 {
		return "unmatched"
	}
	return "partial"
}
