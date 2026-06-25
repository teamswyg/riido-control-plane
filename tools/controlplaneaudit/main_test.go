package main

import "testing"

func TestControlPlaneHighTrafficAuditVerifies(t *testing.T) {
	out := t.TempDir() + "/evidence.json"
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
	got := readEvidence(t, out)
	if got.SurfaceCount < 7 || got.CandidateCount != got.SurfaceCount {
		t.Fatalf("audit coverage = %+v", got)
	}
	assertRequiredCategoriesCovered(t, got.CategoryCounts)
}

func TestControlPlaneHighTrafficAuditRejectsUnsafePprof(t *testing.T) {
	m := loadManifestForTest(t)
	m.PprofCommands = []string{"go tool pprof http://0.0.0.0:6060/debug/pprof/profile"}
	if err := verifyCommands(m); err == nil {
		t.Fatal("expected unsafe pprof command to fail")
	}
}

func TestControlPlaneHighTrafficAuditRequiresManualPressure(t *testing.T) {
	m := loadManifestForTest(t)
	m.ManualPressureCommand = m.LocalPressureCommand
	if err := verifyCommands(m); err == nil {
		t.Fatal("expected missing manual high-concurrency pressure command to fail")
	}
}

func TestControlPlaneHighTrafficAuditRejectsMissingPattern(t *testing.T) {
	m := loadManifestForTest(t)
	m.Surfaces[0].Patterns = []string{"definitely_missing_hot_path_marker"}
	if err := verifySurfaces("../..", m.Surfaces); err == nil {
		t.Fatal("expected missing pattern to fail")
	}
}

func TestControlPlaneHighTrafficAuditRejectsMissingRequiredCategory(t *testing.T) {
	m := loadManifestForTest(t)
	m.RequiredCategories = append(m.RequiredCategories, "missing_traffic_surface")
	if err := verifyRequiredCategories(m.RequiredCategories, m.Surfaces); err == nil {
		t.Fatal("expected missing required category to fail")
	}
}

func assertRequiredCategoriesCovered(t *testing.T, counts map[string]int) {
	t.Helper()
	for _, category := range loadManifestForTest(t).RequiredCategories {
		if counts[category] == 0 {
			t.Fatalf("category %s missing from evidence counts: %+v", category, counts)
		}
	}
}
