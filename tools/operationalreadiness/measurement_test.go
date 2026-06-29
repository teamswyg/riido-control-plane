package main

import "testing"

func TestOperationalReadinessRejectsMissingMeasurement(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].Measurements = nil
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected missing measurement to fail")
	}
}

func TestOperationalReadinessRejectsDuplicateMeasurement(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].Measurements = append(m.Checks[0].Measurements, m.Checks[0].Measurements[0])
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected duplicate measurement to fail")
	}
}
