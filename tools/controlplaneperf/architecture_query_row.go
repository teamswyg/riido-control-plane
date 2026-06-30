package main

func unmatchedArchitectureQueryRow(path string) architectureQueryRow {
	return architectureQueryRow{Path: path, MatchKind: "unmatched"}
}

func exactArchitectureQueryRow(
	path string,
	row architectureFileEvidence,
	components []architectureComponent,
) architectureQueryRow {
	return architectureQueryRow{
		Path:                   path,
		Matched:                true,
		MatchKind:              "exact",
		ComponentIDs:           append([]string(nil), row.ComponentIDs...),
		Components:             architectureQueryComponents(row.ComponentIDs, components),
		HotPathIDs:             append([]string(nil), row.HotPathIDs...),
		PressureDimensions:     append([]string(nil), row.PressureDimensions...),
		ObservabilitySignals:   append([]string(nil), row.ObservabilitySignals...),
		EvidenceRefs:           append([]string(nil), row.EvidenceRefs...),
		TargetVerifierCommands: append([]string(nil), row.TargetVerifierCommands...),
		OptimizationCandidates: append([]string(nil), row.OptimizationCandidates...),
	}
}

func fallbackArchitectureQueryRow(
	path string,
	row architectureFileEvidence,
	components []architectureComponent,
) architectureQueryRow {
	out := exactArchitectureQueryRow(path, row, components)
	out.MatchKind = "directory_fallback"
	return out
}
