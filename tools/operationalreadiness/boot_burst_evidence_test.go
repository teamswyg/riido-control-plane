package main

import (
	"encoding/json"
	"os"
	"testing"
)

const stagingBurstEvidence = "docs/30-architecture/evidence/staging-public-burst-load-evidence-2026-07-02.json"

func TestOperationalReadinessBindsStagingPublicBurstEvidence(t *testing.T) {
	check := readinessCheckByID(t, "boot_burst_capacity")
	if check.Status != "partial" {
		t.Fatalf("boot_burst_capacity status = %q, want partial until cold-start timing evidence", check.Status)
	}
	if !hasMeasurement(check, "staging_public_burst_2026_07_02") {
		t.Fatal("missing staging public burst measurement")
	}
	if !hasEvidenceRef(check, stagingBurstEvidence) {
		t.Fatal("missing staging public burst evidence ref")
	}
}

func TestStagingPublicBurstEvidenceKeepsColdStartPartial(t *testing.T) {
	body, err := os.ReadFile("../../" + stagingBurstEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted bool           `json:"redacted"`
		Findings []burstFinding `json:"findings"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || !hasFinding(evidence.Findings, "cold_start_not_exercised", "partial") {
		t.Fatal("staging burst evidence must remain redacted and preserve cold-start partial finding")
	}
}
