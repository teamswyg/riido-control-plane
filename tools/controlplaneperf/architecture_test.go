package main

import "testing"

func TestControlPlanePerformanceRequiresArchitectureCoverage(t *testing.T) {
	m := loadManifestForTest(t)
	m.ArchitectureComponents = nil
	if err := verifyArchitectureComponents("../..", m); err == nil {
		t.Fatal("expected missing architecture components to fail")
	}
}

func TestControlPlanePerformanceRequiresHotPathCategoryArchitecture(t *testing.T) {
	m := loadManifestForTest(t)
	m.ArchitectureComponents[0].HotPathCategories = nil
	if err := verifyArchitectureComponents("../..", m); err == nil {
		t.Fatal("expected missing hot path category coverage to fail")
	}
}

func TestControlPlanePerformanceRequiresPressureDimensions(t *testing.T) {
	m := loadManifestForTest(t)
	for i := range m.ArchitectureComponents {
		m.ArchitectureComponents[i].PressureDimensions = []string{"latency"}
	}
	if err := verifyArchitectureComponents("../..", m); err == nil {
		t.Fatal("expected missing pressure dimension coverage to fail")
	}
}

func TestControlPlanePerformanceEvidenceCarriesArchitecture(t *testing.T) {
	m := loadManifestForTest(t)
	got := newEvidence(m)
	if got.ArchitectureComponentCount == 0 ||
		len(got.ArchitectureComponents) != got.ArchitectureComponentCount {
		t.Fatalf("architecture evidence missing: %+v", got)
	}
}

func TestControlPlanePerformanceEvidenceCarriesFileArchitectureIndex(t *testing.T) {
	m := loadManifestForTest(t)
	got := newEvidence(m)
	if got.ArchitectureFileCount == 0 ||
		len(got.FileArchitectureIndex) != got.ArchitectureFileCount {
		t.Fatalf("architecture file index missing: %+v", got)
	}
	if got.FileArchitectureIndex[0].Path == "" ||
		len(got.FileArchitectureIndex[0].PressureDimensions) == 0 {
		t.Fatalf("architecture file index row incomplete: %+v", got.FileArchitectureIndex[0])
	}
}
