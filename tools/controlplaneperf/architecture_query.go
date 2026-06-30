package main

const architectureQuerySchema = "riido-control-plane-performance-architecture-query.v1"

func newArchitectureQuery(m manifest, paths []string) architectureQueryEvidence {
	index := architectureRowsByPath(architectureFileIndex(m.ArchitectureComponents, m.HotPaths))
	fallback := architectureFallbackIndex(m.ArchitectureComponents, m.HotPaths)
	rows := make([]architectureQueryRow, 0, len(paths))
	hits := 0
	directHits := 0
	fallbackHits := 0
	for _, path := range paths {
		row := unmatchedArchitectureQueryRow(path)
		if indexed, ok := index[path]; ok {
			row = exactArchitectureQueryRow(path, indexed, m.ArchitectureComponents)
			hits++
			directHits++
		} else if indexed, ok := fallback[architecturePathDir(path)]; ok {
			row = fallbackArchitectureQueryRow(path, indexed, m.ArchitectureComponents)
			hits++
			fallbackHits++
		}
		rows = append(rows, row)
	}
	return architectureQueryEvidence{
		SchemaVersion:    architectureQuerySchema,
		Status:           queryStatus(len(paths), hits),
		QueryCount:       len(paths),
		HitCount:         hits,
		DirectHitCount:   directHits,
		FallbackHitCount: fallbackHits,
		MissCount:        len(paths) - hits,
		Queries:          rows,
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
