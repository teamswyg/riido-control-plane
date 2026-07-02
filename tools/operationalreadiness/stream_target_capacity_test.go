package main

import (
	"encoding/json"
	"os"
	"testing"
)

const streamTargetCapacityAfter = "docs/30-architecture/evidence/" +
	"control-plane-stream-target-capacity-after-2026-07-02.json"

type streamTargetCapacityEvidence struct {
	Redacted                bool               `json:"redacted"`
	ExternalContractChanged bool               `json:"external_contract_changed"`
	EndpointChanged         bool               `json:"endpoint_changed"`
	Observations            []streamTargetCase `json:"observations"`
	Decision                struct {
		Status                      string  `json:"status"`
		AllocationReductionFraction float64 `json:"allocation_reduction_fraction"`
	} `json:"decision"`
}

type streamTargetCase struct {
	Case       string  `json:"case"`
	BytesPerOp float64 `json:"bytes_per_op"`
}

func TestStreamTargetCapacityEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_stream_target_capacity_2026_07_02") {
		t.Fatal("missing stream target capacity measurement")
	}
	if !hasEvidenceRef(check, streamTargetCapacityAfter) {
		t.Fatal("missing stream target capacity evidence ref")
	}
	evidence := loadStreamTargetCapacityEvidence(t)
	if !evidence.Redacted || evidence.ExternalContractChanged || evidence.EndpointChanged {
		t.Fatalf("unsafe stream target evidence flags: %+v", evidence)
	}
	if evidence.Decision.Status != "adopted" || evidence.Decision.AllocationReductionFraction < 0.9 {
		t.Fatalf("insufficient allocation decision: %+v", evidence.Decision)
	}
	if bytesForStreamTargetCase(t, evidence, "current_inactive_heavy") > 128 {
		t.Fatal("inactive-heavy stream target allocation regression")
	}
	if bytesForStreamTargetCase(t, evidence, "legacy_inactive_heavy") < 1024 {
		t.Fatal("legacy baseline should preserve old inactive-heavy allocation")
	}
}

func loadStreamTargetCapacityEvidence(t *testing.T) streamTargetCapacityEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + streamTargetCapacityAfter)
	if err != nil {
		t.Fatal(err)
	}
	var evidence streamTargetCapacityEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

func bytesForStreamTargetCase(t *testing.T, e streamTargetCapacityEvidence, name string) float64 {
	t.Helper()
	for _, observation := range e.Observations {
		if observation.Case == name {
			return observation.BytesPerOp
		}
	}
	t.Fatalf("missing stream target observation %s", name)
	return 0
}
