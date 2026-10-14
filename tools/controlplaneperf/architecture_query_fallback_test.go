package main

import "testing"

func TestControlPlanePerformanceArchitectureQueryRoutesDirectoryFallback(t *testing.T) {
	got := newArchitectureQuery(loadManifestForTest(t), []string{
		"internal/riidoaiserver/server.go",
	})
	if got.Status != "matched" || got.FallbackHitCount != 1 || got.MissCount != 0 {
		t.Fatalf("query fallback status = %+v", got)
	}
	row := got.Queries[0]
	if row.MatchKind != "directory_fallback" || !row.Matched {
		t.Fatalf("query fallback row = %+v", row)
	}
	if len(row.Components) == 0 || len(row.TargetVerifierCommands) == 0 {
		t.Fatalf("query fallback lacks routed context: %+v", row)
	}
}

func TestControlPlanePerformanceArchitectureQueryFallsBackWithinDirectory(t *testing.T) {
	got := newArchitectureQuery(loadManifestForTest(t), []string{
		"internal/riidoaiserver/not-indexed.go",
	})
	if got.Status != "matched" || got.FallbackHitCount != 1 {
		t.Fatalf("query fallback = %+v", got)
	}
	if got.Queries[0].MatchKind != "directory_fallback" {
		t.Fatalf("query fallback match kind = %+v", got.Queries[0])
	}
}
