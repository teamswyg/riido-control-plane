package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestControlPlanePerformanceManifestVerifies(t *testing.T) {
	out := t.TempDir() + "/evidence.json"
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out}); err != nil {
		t.Fatal(err)
	}
	got := readEvidence(t, out)
	if got.HotPathCount < 5 || got.BenchmarkCount < 5 || got.CandidateCount != got.HotPathCount {
		t.Fatalf("evidence coverage = %+v", got)
	}
}

func TestControlPlanePerformanceRejectsMissingBenchmark(t *testing.T) {
	m := loadManifestForTest(t)
	m.HotPaths[0].Benchmarks = []string{"BenchmarkMissingHotPath"}
	if err := verifyAll("../..", m); err == nil {
		t.Fatal("expected missing benchmark to fail")
	}
}

func TestControlPlanePerformanceRequiresLoopbackPprof(t *testing.T) {
	m := loadManifestForTest(t)
	m.PprofCommand = "RIIDO_AI_SERVER_PPROF_ADDR=0.0.0.0:6060"
	if err := verifyCommands(m); err == nil {
		t.Fatal("expected non-loopback pprof command to fail")
	}
}

func TestControlPlanePerformanceRequiresManualPressure(t *testing.T) {
	m := loadManifestForTest(t)
	m.ManualPressureCommand = m.LocalPressureCommand
	if err := verifyCommands(m); err == nil {
		t.Fatal("expected missing manual high-concurrency pressure command to fail")
	}
}

func loadManifestForTest(t *testing.T) manifest {
	t.Helper()
	var m manifest
	if err := readJSON("../../"+defaultManifest, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func readEvidence(t *testing.T, path string) evidence {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}
