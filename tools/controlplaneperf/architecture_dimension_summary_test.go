package main

import "testing"

func TestControlPlanePerformanceEvidenceCarriesPressureDimensions(t *testing.T) {
	got := newEvidence(loadManifestForTest(t))
	byDimension := map[string]pressureDimensionEvidence{}
	for _, row := range got.PressureDimensionSummary {
		byDimension[row.Dimension] = row
	}
	for _, dimension := range []string{"heap_memory", "race_condition", "cpu_busy", "otel_signal"} {
		requirePressureDimensionEvidence(t, byDimension, dimension)
	}
}

func requirePressureDimensionEvidence(
	t *testing.T,
	byDimension map[string]pressureDimensionEvidence,
	dimension string,
) {
	t.Helper()
	row, ok := byDimension[dimension]
	if !ok {
		t.Fatalf("pressure dimension %q missing from evidence", dimension)
	}
	if len(row.Files) == 0 || len(row.ObservabilitySignals) == 0 ||
		len(row.TargetVerifierCommands) == 0 {
		t.Fatalf("pressure dimension row incomplete: %+v", row)
	}
}
