package main

import "testing"

func TestOperationalReadinessBindsClientSurfaceAlarmApplyEvidence(t *testing.T) {
	check := readinessCheckByID(t, "otel_xray_client_surface")
	if check.Status != "partial" {
		t.Fatalf("client surface alarm must remain partial until live apply evidence: %s", check.Status)
	}
	if check.NextArtifact != "client_surface_alarm_plan_apply_evidence" {
		t.Fatalf("next artifact = %q", check.NextArtifact)
	}
	if !hasMeasurement(check, "client_surface_alarm_topology_pr") {
		t.Fatal("missing merged infra alarm topology measurement")
	}
	if !hasEvidenceRef(check, "github:https://github.com/teamswyg/riido-infra/pull/113") {
		t.Fatal("missing riido-infra PR #113 evidence ref")
	}
}

func readinessCheckByID(t *testing.T, id string) readinessCheck {
	t.Helper()
	for _, check := range loadManifestForTest(t).Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("missing readiness check %s", id)
	return readinessCheck{}
}

func hasMeasurement(check readinessCheck, id string) bool {
	for _, measurement := range check.Measurements {
		if measurement.ID == id {
			return true
		}
	}
	return false
}

func hasEvidenceRef(check readinessCheck, path string) bool {
	for _, ref := range check.EvidenceRefs {
		if ref.Path == path {
			return true
		}
	}
	return false
}
