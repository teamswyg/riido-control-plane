package main

import (
	"os"
	"strings"
	"testing"
)

func verifySourceContractsManifestProvenance(t *testing.T, sourceStabilizedBy, projectionStabilizedBy []string, docPath string) {
	t.Helper()
	if len(sourceStabilizedBy) != len(expectedFigmaSourceProvenance) {
		t.Fatalf("mirrored source coverage stabilized_by = %d entries, want %d: %+v", len(sourceStabilizedBy), len(expectedFigmaSourceProvenance), sourceStabilizedBy)
	}
	for i := range expectedFigmaSourceProvenance {
		if sourceStabilizedBy[i] != expectedFigmaSourceProvenance[i] {
			t.Fatalf("mirrored source coverage stabilized_by[%d] = %q, want %q; full list = %+v", i, sourceStabilizedBy[i], expectedFigmaSourceProvenance[i], sourceStabilizedBy)
		}
	}
	if len(projectionStabilizedBy) != len(sourceStabilizedBy) {
		t.Fatalf("source_contracts_manifest.stabilized_by = %d entries, mirrored source has %d: projection=%+v source=%+v", len(projectionStabilizedBy), len(sourceStabilizedBy), projectionStabilizedBy, sourceStabilizedBy)
	}
	for i := range sourceStabilizedBy {
		if projectionStabilizedBy[i] != sourceStabilizedBy[i] {
			t.Fatalf("source_contracts_manifest.stabilized_by[%d] = %q, mirrored source stabilized_by[%d] = %q; projection=%+v source=%+v", i, projectionStabilizedBy[i], i, sourceStabilizedBy[i], projectionStabilizedBy, sourceStabilizedBy)
		}
	}
	verifySourceProvenanceDoc(t, docPath)
}

func verifySourceProvenanceDoc(t *testing.T, docPath string) {
	t.Helper()
	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read projection doc for source provenance: %v", err)
	}
	docText := string(data)
	for _, pr := range expectedFigmaSourceProvenance {
		if !strings.Contains(docText, pr) {
			t.Fatalf("projection doc must mention full upstream contracts provenance %q", pr)
		}
	}
	if !strings.Contains(docText, "full upstream coverage provenance") ||
		!strings.Contains(docText, "limitation-local provenance") {
		t.Fatalf("projection doc must distinguish full upstream coverage provenance from limitation-local provenance")
	}
}
