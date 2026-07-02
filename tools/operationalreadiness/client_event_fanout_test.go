package main

import (
	"encoding/json"
	"os"
	"testing"
)

const clientEventFanoutAfter = "docs/30-architecture/evidence/" +
	"control-plane-client-event-fanout-after-2026-07-02.json"

type clientEventFanoutEvidence struct {
	Redacted                bool           `json:"redacted"`
	ExternalContractChanged bool           `json:"external_contract_changed"`
	EndpointChanged         bool           `json:"endpoint_changed"`
	Before                  fanoutCapacity `json:"before_capacity"`
	After                   fanoutCapacity `json:"after_capacity"`
	Decision                fanoutDecision `json:"decision"`
}

type fanoutCapacity struct {
	AllocBytesPerOp float64 `json:"alloc_bytes_per_op"`
	CPUSecondsPerOp float64 `json:"cpu_seconds_per_op"`
	GoroutineDelta  int     `json:"goroutine_delta"`
	ErrorFree       bool    `json:"error_free"`
}

type fanoutDecision struct {
	Status                      string  `json:"status"`
	AllocationReductionFraction float64 `json:"allocation_reduction_fraction"`
	CPUSecondsReductionFraction float64 `json:"cpu_seconds_per_op_reduction_fraction"`
}

func TestClientEventFanoutEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_client_event_fanout_2026_07_02") {
		t.Fatal("missing client event fanout measurement")
	}
	if !hasEvidenceRef(check, clientEventFanoutAfter) {
		t.Fatal("missing client event fanout evidence ref")
	}
	evidence := loadClientEventFanoutEvidence(t)
	if !evidence.Redacted || evidence.ExternalContractChanged || evidence.EndpointChanged {
		t.Fatalf("unsafe fanout evidence flags: %+v", evidence)
	}
	if !evidence.After.ErrorFree || evidence.After.GoroutineDelta != 0 {
		t.Fatalf("fanout after run is not clean: %+v", evidence.After)
	}
	if evidence.After.AllocBytesPerOp >= evidence.Before.AllocBytesPerOp {
		t.Fatalf("fanout allocation did not improve: %+v", evidence)
	}
	if evidence.Decision.Status != "adopted" ||
		evidence.Decision.AllocationReductionFraction < 0.1 ||
		evidence.Decision.CPUSecondsReductionFraction < 0.1 {
		t.Fatalf("weak fanout decision: %+v", evidence.Decision)
	}
}

func loadClientEventFanoutEvidence(t *testing.T) clientEventFanoutEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + clientEventFanoutAfter)
	if err != nil {
		t.Fatal(err)
	}
	var evidence clientEventFanoutEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}
