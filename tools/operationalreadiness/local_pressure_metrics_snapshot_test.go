package main

import (
	"encoding/json"
	"os"
	"testing"
)

const (
	httpMetricsSnapshotEvidence  = "docs/30-architecture/evidence/control-plane-http-metrics-snapshot-optimization-2026-07-03.json"
	storeMetricsSnapshotEvidence = "docs/30-architecture/evidence/control-plane-store-metrics-snapshot-optimization-2026-07-03.json"
)

func TestLocalPressureMetricsSnapshotEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	assertMetricsSnapshotEvidence(t, check, "local_pressure_http_metrics_snapshot_2026_07_03", httpMetricsSnapshotEvidence, 20, 25, 25)
	assertMetricsSnapshotEvidence(t, check, "local_pressure_store_metrics_snapshot_2026_07_03", storeMetricsSnapshotEvidence, 35, 80, 50)
}

func assertMetricsSnapshotEvidence(t *testing.T, check readinessCheck, measurementID, path string, minLatency, minBytes, minAllocs float64) {
	t.Helper()
	if !hasMeasurement(check, measurementID) {
		t.Fatalf("missing metrics snapshot measurement %s", measurementID)
	}
	if !hasEvidenceRef(check, path) {
		t.Fatalf("missing metrics snapshot evidence ref %s", path)
	}
	evidence := loadMetricsSnapshotEvidence(t, path)
	if !evidence.Redacted || evidence.Change.ContractImpact != "none" {
		t.Fatalf("metrics snapshot evidence must be redacted and contract-safe: %+v", evidence)
	}
	if evidence.Improvement.MedianLatencyReductionPercent < minLatency ||
		evidence.Improvement.BytesPerOpReductionPercent < minBytes ||
		evidence.Improvement.AllocsPerOpReductionPercent < minAllocs {
		t.Fatalf("metrics snapshot improvement too weak: %+v", evidence.Improvement)
	}
}

func loadMetricsSnapshotEvidence(t *testing.T, path string) metricsSnapshotEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence metricsSnapshotEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

type metricsSnapshotEvidence struct {
	Redacted bool `json:"redacted"`
	Change   struct {
		ContractImpact string `json:"contract_impact"`
	} `json:"change"`
	Improvement struct {
		MedianLatencyReductionPercent float64 `json:"median_latency_reduction_percent"`
		BytesPerOpReductionPercent    float64 `json:"bytes_per_op_reduction_percent"`
		AllocsPerOpReductionPercent   float64 `json:"allocs_per_op_reduction_percent"`
	} `json:"improvement"`
}
