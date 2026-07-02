package main

import (
	"encoding/json"
	"os"
	"testing"
)

const clientSurfaceAlarmPreflightEvidence = "docs/30-architecture/evidence/client-surface-alarm-preflight-development-2026-07-02.json"

func TestOperationalReadinessBindsClientSurfaceAlarmPreflightEvidence(t *testing.T) {
	check := readinessCheckByID(t, "otel_xray_client_surface")
	if check.Status != "partial" {
		t.Fatalf("client surface alarm must remain partial until live apply evidence: %s", check.Status)
	}
	if !hasMeasurement(check, "client_surface_alarm_preflight_development_2026_07_02") {
		t.Fatal("missing development alarm preflight measurement")
	}
	if !hasEvidenceRef(check, clientSurfaceAlarmPreflightEvidence) {
		t.Fatal("missing development alarm preflight evidence ref")
	}
}

func TestClientSurfaceAlarmPreflightEvidenceStaysRedacted(t *testing.T) {
	body, err := os.ReadFile("../../" + clientSurfaceAlarmPreflightEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted  bool `json:"redacted"`
		Preflight struct {
			Status               string   `json:"status"`
			MissingRequiredInput []string `json:"missing_required_input_kinds"`
			RawPlanLogIncluded   bool     `json:"raw_plan_log_included"`
			TFVarsValuesIncluded bool     `json:"tfvars_values_included"`
			SecretsIncluded      bool     `json:"secrets_included"`
			Passed               bool     `json:"passed"`
		} `json:"preflight"`
		Decision struct {
			Status string `json:"status"`
		} `json:"decision"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || evidence.Decision.Status != "partial" {
		t.Fatalf("unexpected evidence state: %+v", evidence.Decision)
	}
	if evidence.Preflight.Status != "blocked_missing_operator_tfvars" || len(evidence.Preflight.MissingRequiredInput) == 0 {
		t.Fatalf("unexpected preflight block: %+v", evidence.Preflight)
	}
	if evidence.Preflight.RawPlanLogIncluded || evidence.Preflight.TFVarsValuesIncluded || evidence.Preflight.SecretsIncluded {
		t.Fatal("preflight evidence must not include raw plan logs, tfvars values, or secrets")
	}
	if !evidence.Preflight.Passed {
		t.Fatal("redacted preflight evidence should pass even when live apply is blocked")
	}
}
