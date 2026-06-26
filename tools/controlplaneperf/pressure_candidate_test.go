package main

import "testing"

func TestControlPlanePerformanceRequiresPressureCandidateOutput(t *testing.T) {
	m := loadManifestForTest(t)
	m.LocalPressureCommand = "go run ./tools/controlplanepressure -duration 500ms"
	if err := verifyCommands(m); err == nil {
		t.Fatal("expected missing pressure candidate output to fail")
	}
}

func TestControlPlanePerformanceRequiresPressureCandidateSource(t *testing.T) {
	m := loadManifestForTest(t)
	m.Sources = nil
	if err := verifyPressureSources(m); err == nil {
		t.Fatal("expected missing pressure candidate source to fail")
	}
}
