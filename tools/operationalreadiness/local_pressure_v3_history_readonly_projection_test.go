package main

import "testing"

const localPressureV3HistoryReadonlyProjectionAfter = "docs/30-architecture/evidence/control-plane-v3-history-readonly-projection-after-2026-07-03.json"

func TestLocalPressureV3HistoryReadonlyProjectionEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_v3_history_readonly_projection_2026_07_03") {
		t.Fatal("missing v3 history readonly projection measurement")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryReadonlyProjectionAfter) {
		t.Fatal("missing v3 history readonly projection evidence ref")
	}
	before := loadPressureEvidence(t, localPressureProgressMessageIDStackAfter)
	after := loadPressureEvidence(t, localPressureV3HistoryReadonlyProjectionAfter)
	assertCleanPressureEvidence(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.30)
	assertAllocationReduced(t, before, after, "http_endpoint_threads_v3", 0.10)
}
