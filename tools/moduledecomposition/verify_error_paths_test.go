package main

import (
	"strings"
	"testing"
)

func TestVerifyManifestShapeReportsMissingRequiredFields(t *testing.T) {
	t.Parallel()
	if err := verifyManifestShape(manifest{}); err == nil ||
		!strings.Contains(err.Error(), "schema_version") {
		t.Fatalf("expected required field error, got %v", err)
	}
	m := manifest{SchemaVersion: "v", ID: "id", Title: "title", GeneratedDoc: "doc.md"}
	if err := verifyManifestShape(m); err == nil ||
		!strings.Contains(err.Error(), "module_path") {
		t.Fatalf("expected module path error, got %v", err)
	}
}

func TestVerifyManifestShapeReportsLineBudgetAndLoopErrors(t *testing.T) {
	t.Parallel()
	m := validShapeManifest()
	m.FileLineBudget.TargetLines = -1
	if err := verifyManifestShape(m); err == nil ||
		!strings.Contains(err.Error(), "non-negative") {
		t.Fatalf("expected budget error, got %v", err)
	}
	m = validShapeManifest()
	m.FileLineBudget.HotspotLimits = []lineBudgetHotspotLimit{{MaxFiles: -1}}
	if err := verifyManifestShape(m); err == nil ||
		!strings.Contains(err.Error(), "hotspot limits") {
		t.Fatalf("expected hotspot limit error, got %v", err)
	}
	m = validShapeManifest()
	m.Loop.Evaluate = ""
	if err := verifyManifestShape(m); err == nil ||
		!strings.Contains(err.Error(), "loop must define") {
		t.Fatalf("expected loop error, got %v", err)
	}
}

func validShapeManifest() manifest {
	return manifest{
		SchemaVersion: "v",
		ID:            "id",
		Title:         "title",
		GeneratedDoc:  "doc.md",
		ModulePath:    "example.com/fixture",
		SourceRoots:   []string{"cmd"},
		FileLineBudget: fileLineBudget{
			HotspotLimits: []lineBudgetHotspotLimit{{Path: "cmd", MaxFiles: 1}},
		},
		Packages: []packageEntry{{Path: "cmd/app", Kind: "runtime", Role: "entrypoint", MustNotOwn: "domain"}},
		Loop: evidenceLoop{
			Observation:   "o",
			Hypothesis:    "h",
			Execute:       "x",
			Evaluate:      "e",
			Retrospective: "r",
		},
	}
}
