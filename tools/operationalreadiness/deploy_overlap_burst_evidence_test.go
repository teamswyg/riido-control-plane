package main

import (
	"encoding/json"
	"os"
	"testing"
)

const deployOverlapBurstEvidence = "docs/30-architecture/evidence/staging-deploy-overlap-burst-evidence-2026-07-02.json"

func TestOperationalReadinessBindsDeployOverlapBurstEvidence(t *testing.T) {
	check := readinessCheckByID(t, "boot_burst_capacity")
	if !hasMeasurement(check, "staging_deploy_overlap_burst_2026_07_02") {
		t.Fatal("missing staging deploy-overlap burst measurement")
	}
	if !hasEvidenceRef(check, deployOverlapBurstEvidence) {
		t.Fatal("missing staging deploy-overlap burst evidence ref")
	}
}

func TestDeployOverlapBurstEvidenceKeepsAttributionPartial(t *testing.T) {
	body, err := os.ReadFile("../../" + deployOverlapBurstEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted bool `json:"redacted"`
		Runs     struct {
			Load struct {
				Total    int `json:"total"`
				Success  int `json:"success"`
				Failures int `json:"failures"`
			} `json:"load"`
		} `json:"runs"`
		Overlap struct {
			CoversDeploy bool `json:"load_covered_build_register_deploy_wait_and_smoke"`
			Seconds      int  `json:"runtime_deploy_wait_overlap_seconds"`
		} `json:"overlap"`
		Findings []burstFinding `json:"findings"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || !evidence.Overlap.CoversDeploy {
		t.Fatal("deploy-overlap evidence must be redacted and cover the deploy window")
	}
	if evidence.Overlap.Seconds <= 0 {
		t.Fatal("deploy-overlap evidence must record positive runtime deploy wait overlap")
	}
	if evidence.Runs.Load.Total == 0 || evidence.Runs.Load.Success != evidence.Runs.Load.Total {
		t.Fatalf("unexpected load result: %+v", evidence.Runs.Load)
	}
	if evidence.Runs.Load.Failures != 0 {
		t.Fatalf("load failures = %d, want 0", evidence.Runs.Load.Failures)
	}
	if !hasFinding(evidence.Findings, "cold_start_target_attribution_missing", "partial") {
		t.Fatal("deploy-overlap evidence must preserve target attribution partial")
	}
}
