package main

import (
	"encoding/json"
	"os"
	"testing"
)

const localPressureFollowup = "docs/30-architecture/evidence/control-plane-local-pressure-followup-2026-07-02.json"

func TestLocalPressureFollowupEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_followup_2026_07_02") {
		t.Fatal("missing local pressure follow-up measurement")
	}
	if !hasEvidenceRef(check, localPressureFollowup) {
		t.Fatal("missing local pressure follow-up evidence ref")
	}
	evidence := loadLocalPressureFollowup(t)
	if len(evidence.Runs) != 27 || len(evidence.Capacity) != 9 {
		t.Fatalf("unexpected pressure shape: runs=%d capacity=%d", len(evidence.Runs), len(evidence.Capacity))
	}
	for _, run := range evidence.Runs {
		if run.Errors != 0 || run.Resources.Goroutines != 0 {
			t.Fatalf("pressure run not clean: %+v", run)
		}
	}
	if !hasPressureFinding(evidence.Findings, "allocation_hotspot", "http_endpoint_threads_v3") {
		t.Fatal("missing v3 endpoint allocation finding")
	}
}

func loadLocalPressureFollowup(t *testing.T) pressureEvidence {
	t.Helper()
	return loadPressureEvidence(t, localPressureFollowup)
}

func loadPressureEvidence(t *testing.T, path string) pressureEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence pressureEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

func hasPressureFinding(findings []pressureFinding, id, scenario string) bool {
	for _, finding := range findings {
		if finding.ID == id && finding.Scenario == scenario && finding.Value > 0 {
			return true
		}
	}
	return false
}
