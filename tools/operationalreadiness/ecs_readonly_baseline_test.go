package main

import (
	"encoding/json"
	"os"
	"testing"
)

const ecsBaselineEvidence = "docs/30-architecture/evidence/ecs-testnet-recovery-scale-readonly-baseline-2026-07-02.json"

func TestOperationalReadinessBindsECSReadonlyBaseline(t *testing.T) {
	for _, tc := range []struct {
		checkID       string
		measurementID string
	}{
		{"server_crash_recovery", "ecs_testnet_recovery_readonly_baseline_2026_07_02"},
		{"scale_out_recovery", "ecs_testnet_scale_readonly_baseline_2026_07_02"},
	} {
		check := readinessCheckByID(t, tc.checkID)
		if check.Status != "partial" {
			t.Fatalf("%s status = %q, want partial until chaos timing evidence", tc.checkID, check.Status)
		}
		if !hasMeasurement(check, tc.measurementID) {
			t.Fatalf("%s missing %s measurement", tc.checkID, tc.measurementID)
		}
		if !hasEvidenceRef(check, ecsBaselineEvidence) {
			t.Fatalf("%s missing ECS baseline evidence ref", tc.checkID)
		}
	}
}

func TestECSReadonlyBaselineEvidenceIsRedacted(t *testing.T) {
	body, err := os.ReadFile("../../" + ecsBaselineEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted    bool `json:"redacted"`
		Service     any  `json:"service"`
		Autoscaling any  `json:"autoscaling"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || evidence.Service == nil || evidence.Autoscaling == nil {
		t.Fatal("ECS baseline evidence must be redacted and include service/autoscaling summaries")
	}
}
