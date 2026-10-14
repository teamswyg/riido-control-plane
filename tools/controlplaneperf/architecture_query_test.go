package main

import "testing"

func TestControlPlanePerformanceArchitectureQueryRoutesHotPath(t *testing.T) {
	out := t.TempDir() + "/query.json"
	err := run(options{
		Repo:                 "../..",
		Manifest:             defaultManifest,
		ArchitectureQueryOut: out,
		ArchitecturePaths: []string{
			"internal/riidoaiserver/http_transaction_metrics_observe.go",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := readArchitectureQuery(t, out)
	if got.Status != "matched" || got.HitCount != 1 || got.MissCount != 0 {
		t.Fatalf("query status = %+v", got)
	}
	row := got.Queries[0]
	if row.MatchKind != "exact" || got.DirectHitCount != 1 || got.FallbackHitCount != 0 {
		t.Fatalf("query match kind = %+v", got)
	}
	if len(row.TargetVerifierCommands) == 0 || len(row.ObservabilitySignals) == 0 {
		t.Fatalf("query row lacks routing evidence: %+v", row)
	}
	if len(row.Components) == 0 || row.Components[0].Role == "" {
		t.Fatalf("query row lacks component summary: %+v", row)
	}
}

func TestControlPlanePerformanceArchitectureQueryReportsMisses(t *testing.T) {
	got := newArchitectureQuery(loadManifestForTest(t), []string{
		"cmd/not-indexed.go",
	})
	if got.Status != "unmatched" || got.MissCount != 1 ||
		got.Queries[0].Matched || got.Queries[0].MatchKind != "unmatched" {
		t.Fatalf("query miss = %+v", got)
	}
}

func TestControlPlanePerformanceArchitectureQueryRequiresPath(t *testing.T) {
	err := run(options{
		Repo:                 "../..",
		Manifest:             defaultManifest,
		ArchitectureQueryOut: t.TempDir() + "/query.json",
	})
	if err == nil {
		t.Fatal("expected query output without paths to fail")
	}
}

func TestControlPlanePerformanceRequiresArchitectureQueryArtifact(t *testing.T) {
	m := loadManifestForTest(t)
	m.ArchitectureQueryArtifact = "missing-query-artifact"
	if err := verifyWorkflow("../..", m); err == nil {
		t.Fatal("expected missing architecture query artifact upload to fail")
	}
}
