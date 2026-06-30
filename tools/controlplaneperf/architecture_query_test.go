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
	if len(row.TargetVerifierCommands) == 0 || len(row.ObservabilitySignals) == 0 {
		t.Fatalf("query row lacks routing evidence: %+v", row)
	}
}

func TestControlPlanePerformanceArchitectureQueryReportsMisses(t *testing.T) {
	got := newArchitectureQuery(loadManifestForTest(t), []string{
		"internal/riidoaiserver/not-indexed.go",
	})
	if got.Status != "unmatched" || got.MissCount != 1 || got.Queries[0].Matched {
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
