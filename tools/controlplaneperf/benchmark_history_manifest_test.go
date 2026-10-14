package main

import "testing"

func TestControlPlanePerformanceRejectsMissingBenchmarkHistory(t *testing.T) {
	m := loadManifestForTest(t)
	m.BenchmarkHistory = ""
	if err := verifyAll("../..", m); err == nil {
		t.Fatal("expected missing benchmark history to fail")
	}
}

func TestControlPlanePerformanceRejectsWeakBenchmarkHistoryCommand(t *testing.T) {
	m := loadManifestForTest(t)
	m.BenchmarkHistoryCommand = "go run ./tools/controlplaneperf"
	if err := verifyAll("../..", m); err == nil {
		t.Fatal("expected weak benchmark history command to fail")
	}
}
