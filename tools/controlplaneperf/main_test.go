package main

import "testing"

func TestControlPlanePerformanceManifestVerifies(t *testing.T) {
	out := t.TempDir() + "/evidence.json"
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out}); err != nil {
		t.Fatal(err)
	}
	got := readEvidence(t, out)
	if got.HotPathCount < 5 || got.BenchmarkCount < 5 || got.CandidateCount != got.HotPathCount {
		t.Fatalf("evidence coverage = %+v", got)
	}
	if got.RaceArtifact == "" || got.RaceCommand == "" {
		t.Fatalf("race evidence contract missing: %+v", got)
	}
	if got.AssertionCount == 0 || len(got.Assertions) != got.AssertionCount {
		t.Fatalf("performance assertions missing from evidence: %+v", got)
	}
	if len(got.Sources) == 0 || got.Sources[0].HarnessLoop == "" {
		t.Fatalf("performance source context missing from evidence: %+v", got.Sources)
	}
	if got.Loop.Observation == "" || got.Loop.Evaluate == "" {
		t.Fatalf("performance loop context missing from evidence: %+v", got.Loop)
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

func TestControlPlanePerformanceRejectsMissingLocalPressureScenario(t *testing.T) {
	m := loadManifestForTest(t)
	m.LocalPressureScenarios = append(m.LocalPressureScenarios, "missing_pressure_scenario")
	if err := verifyLocalPressureScenarios("../..", m.LocalPressureScenarios); err == nil {
		t.Fatal("expected missing local pressure scenario to fail")
	}
}
