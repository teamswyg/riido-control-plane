package main

import (
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/requirements"
)

func assertGeneratedSmokeMatrixGolden(t *testing.T, got evidence) {
	t.Helper()
	if got.SchemaVersion != requirements.EvidenceSchema ||
		got.ID != requirements.ExpectedID ||
		got.Status != "verified" ||
		!got.MatrixParity ||
		!got.MatrixSorted {
		t.Fatalf("unexpected evidence identity/status: %+v", got)
	}
	if got.Counts.Total != 57 || got.Counts.V1 != 26 || got.Counts.V2 != 31 {
		t.Fatalf("operation counts drifted: %+v", got.Counts)
	}
	if got.SourceChecks != 4 || got.EvidenceTests != 2 {
		t.Fatalf("evidence counters drifted: %+v", got)
	}
	if got.Loop.Observation == "" ||
		got.Loop.Hypothesis == "" ||
		got.Loop.Execute == "" ||
		got.Loop.Evaluate == "" ||
		got.Loop.Retrospective == "" {
		t.Fatalf("loop evidence must stay populated: %+v", got.Loop)
	}
}
