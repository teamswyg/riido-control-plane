package main

import (
	"encoding/json"
	"os"
	"testing"
)

const localPressureV3HistoryQueuedSuppressionFastPath = "docs/30-architecture/evidence/" +
	"control-plane-v3-history-queued-suppression-fastpath-2026-07-03.json"

func TestLocalPressureV3HistoryQueuedSuppressionFastPathEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_v3_history_queued_suppression_fastpath_2026_07_03") {
		t.Fatal("missing v3 history queued suppression fast path measurement")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryQueuedSuppressionFastPath) {
		t.Fatal("missing v3 history queued suppression fast path evidence ref")
	}
	evidence := loadQueuedSuppressionFastPathEvidence(t)
	if evidence.Change.EndpointContractChanged || evidence.Change.DurableQueueChanged {
		t.Fatalf("queued suppression fast path must not change contracts: %+v", evidence.Change)
	}
	if evidence.Decision.LatencyReductionPercent < 50 ||
		evidence.Decision.BytesPerOpReductionPercent < 20 ||
		evidence.Decision.AllocsPerOpReductionPercent < 15 {
		t.Fatalf("queued suppression fast path evidence is too weak: %+v", evidence.Decision)
	}
}

func loadQueuedSuppressionFastPathEvidence(t *testing.T) queuedSuppressionFastPathEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + localPressureV3HistoryQueuedSuppressionFastPath)
	if err != nil {
		t.Fatal(err)
	}
	var evidence queuedSuppressionFastPathEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

type queuedSuppressionFastPathEvidence struct {
	Change struct {
		EndpointContractChanged bool `json:"endpoint_contract_changed"`
		DurableQueueChanged     bool `json:"durable_queue_changed"`
	} `json:"change"`
	Decision struct {
		LatencyReductionPercent     float64 `json:"latency_reduction_percent"`
		BytesPerOpReductionPercent  float64 `json:"bytes_per_op_reduction_percent"`
		AllocsPerOpReductionPercent float64 `json:"allocs_per_op_reduction_percent"`
	} `json:"decision"`
}
