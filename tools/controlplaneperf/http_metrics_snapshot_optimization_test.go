package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestHTTPMetricsSnapshotOptimizationEvidence(t *testing.T) {
	path := "../../docs/30-architecture/evidence/control-plane-http-metrics-snapshot-optimization-2026-07-03.json"
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		SchemaVersion string `json:"schema_version"`
		Redacted      bool   `json:"redacted"`
		Change        struct {
			File           string `json:"file"`
			ContractImpact string `json:"contract_impact"`
		} `json:"change"`
		Improvement struct {
			MedianLatencyReductionPercent float64 `json:"median_latency_reduction_percent"`
			BytesPerOpReductionPercent    float64 `json:"bytes_per_op_reduction_percent"`
			AllocsPerOpReductionPercent   float64 `json:"allocs_per_op_reduction_percent"`
		} `json:"improvement"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.SchemaVersion == "" || !evidence.Redacted {
		t.Fatalf("invalid optimization evidence metadata: %+v", evidence)
	}
	if evidence.Change.File != "internal/riidoaiserver/http_transaction_metrics_snapshot.go" ||
		evidence.Change.ContractImpact != "none" {
		t.Fatalf("optimization evidence must bind exact no-contract-change file: %+v", evidence.Change)
	}
	if evidence.Improvement.MedianLatencyReductionPercent < 20 ||
		evidence.Improvement.BytesPerOpReductionPercent < 25 ||
		evidence.Improvement.AllocsPerOpReductionPercent < 25 {
		t.Fatalf("optimization evidence is too weak: %+v", evidence.Improvement)
	}
}
