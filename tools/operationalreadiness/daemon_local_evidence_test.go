package main

import "testing"

func TestOperationalReadinessBindsDaemonLocalEvidence(t *testing.T) {
	m := loadManifestForTest(t)
	for _, id := range []string{
		"daemon_network_disconnect_waiting",
		"single_pc_agent_limit",
		"all_servers_down_daemon_behavior",
	} {
		check := findReadinessCheck(t, m, id)
		if check.Status != "partial" {
			t.Fatalf("%s status = %q, want partial until release evidence", id, check.Status)
		}
		if !hasMeasurementEvidence(check, "daemon-local-network-capacity-evidence-2026-07-02.json") {
			t.Fatalf("%s missing daemon local evidence measurement", id)
		}
		if !hasMeasurementEvidence(check, "daemon-reconnect-runtimeactor-evidence-2026-07-02.json") {
			t.Fatalf("%s missing daemon reconnect/runtimeactor evidence measurement", id)
		}
	}
}

func hasMeasurementEvidence(check readinessCheck, want string) bool {
	for _, measurement := range check.Measurements {
		if measurement.EvidenceRef == "docs/30-architecture/evidence/"+want {
			return true
		}
	}
	return false
}
