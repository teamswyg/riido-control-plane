package main

import "testing"

func TestOperationalReadinessRejectsCoveredVisualWithoutScreenshot(t *testing.T) {
	m := loadManifestForTest(t)
	check := findReadinessCheck(t, m, "staging_client_p0_visual_retest")
	check.Status = "covered"
	check.Measurements = filterOutMeasurementKind(check.Measurements, "screenshot")
	check.Measurements = append(check.Measurements, measurement{
		ID: "api_state", Kind: "artifact", Signal: "Server state is green.",
	})
	replaceReadinessCheck(t, &m, check)
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected covered visual QA without screenshot evidence to fail")
	}
}

func filterOutMeasurementKind(measurements []measurement, kind string) []measurement {
	filtered := measurements[:0]
	for _, measurement := range measurements {
		if measurement.Kind != kind {
			filtered = append(filtered, measurement)
		}
	}
	return filtered
}

func TestOperationalReadinessAcceptsCoveredVisualWithScreenshot(t *testing.T) {
	m := loadManifestForTest(t)
	check := findReadinessCheck(t, m, "staging_client_p0_visual_retest")
	check.Status = "covered"
	check.Measurements = append(check.Measurements, measurement{
		ID: "staging_client_screen", Kind: "screenshot", Signal: "Redacted staging client visual QA screenshot.",
	})
	replaceReadinessCheck(t, &m, check)
	if err := verifyChecks("../..", m); err != nil {
		t.Fatal(err)
	}
}

func TestOperationalReadinessRejectsVisualCommandWithoutIssueDetails(t *testing.T) {
	m := loadManifestForTest(t)
	check := findReadinessCheck(t, m, "staging_client_p0_visual_retest")
	check.NextCommand = "gh issue view 711 --repo teamswyg/riido-control-plane --json url,title,state"
	replaceReadinessCheck(t, &m, check)
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected visual QA command without issue body and comments to fail")
	}
}

func findReadinessCheck(t *testing.T, m manifest, id string) readinessCheck {
	t.Helper()
	for _, check := range m.Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("missing readiness check %s", id)
	return readinessCheck{}
}

func replaceReadinessCheck(t *testing.T, m *manifest, replacement readinessCheck) {
	t.Helper()
	for i, check := range m.Checks {
		if check.ID == replacement.ID {
			m.Checks[i] = replacement
			return
		}
	}
	t.Fatalf("missing readiness check %s", replacement.ID)
}
