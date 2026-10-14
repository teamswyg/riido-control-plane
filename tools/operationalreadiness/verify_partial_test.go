package main

import "testing"

func TestOperationalReadinessRejectsManualOnlyPartial(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].Status = "partial"
	m.Checks[0].Measurements = []measurement{{
		ID: "manual_note", Kind: "manual", Signal: "A human noted the issue.",
	}}
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected manual-only partial evidence to fail")
	}
}
