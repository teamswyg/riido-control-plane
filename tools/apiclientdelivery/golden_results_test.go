package main

import "testing"

func assertAPIClientDeliveryGolden(t *testing.T, got evidence) {
	t.Helper()
	if got.SchemaVersion != "riido-control-plane-api-client-delivery-evidence.v1" {
		t.Fatalf("schema_version = %q", got.SchemaVersion)
	}
	if got.ID != "api-client-delivery" || got.Status != "verified" {
		t.Fatalf("identity/status = %q/%q", got.ID, got.Status)
	}
	wantCounts := []struct {
		name string
		got  int
		want int
	}{
		{"source_manifests", got.Result.SourceManifests, 4},
		{"owners", got.Result.Owners, 3},
		{"figma_contexts", got.Result.FigmaContexts, 9},
		{"source_checks", got.Result.SourceChecks, 7},
		{"phrase_checks", got.Result.PhraseChecks, 29},
		{"forbidden_checks", got.Result.ForbiddenChecks, 3},
		{"risk_tests", got.Result.RiskTests, 14},
	}
	for _, count := range wantCounts {
		if count.got != count.want {
			t.Fatalf("%s = %d, want %d", count.name, count.got, count.want)
		}
	}
	if got.Workflow != ".github/workflows/api-client-delivery.yml" {
		t.Fatalf("workflow = %q", got.Workflow)
	}
	if got.GeneratedDoc != "docs/30-architecture/api-client-delivery.md" {
		t.Fatalf("generated_doc = %q", got.GeneratedDoc)
	}
	if got.Delivery.Workflow != ".github/workflows/generated-client-delivery.yml" {
		t.Fatalf("delivery workflow = %q", got.Delivery.Workflow)
	}
	if got.Loop.Observation == "" || got.Loop.Retrospective == "" {
		t.Fatalf("loop must be populated: %+v", got.Loop)
	}
}
