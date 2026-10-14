package main

import "testing"

func TestControlPlanePerformanceRequiresHotPathFileArchitecture(t *testing.T) {
	m := loadManifestForTest(t)
	m.ArchitectureComponents[0].Files = []string{"internal/riidoaiserver/server_handler.go"}
	if err := verifyArchitectureComponents("../..", m); err == nil {
		t.Fatal("expected missing hot path file architecture coverage to fail")
	}
}

func TestControlPlanePerformanceRejectsDuplicateArchitectureFile(t *testing.T) {
	m := loadManifestForTest(t)
	m.ArchitectureComponents[0].Files = append(
		m.ArchitectureComponents[0].Files,
		m.ArchitectureComponents[0].Files[0],
	)
	if err := verifyArchitectureComponents("../..", m); err == nil {
		t.Fatal("expected duplicate architecture file to fail")
	}
}
