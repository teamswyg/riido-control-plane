package main

import "testing"

const localPressureV3HistoryMessageIDAfter = "docs/30-architecture/evidence/control-plane-v3-history-message-id-allocation-after-2026-07-02.json"

func TestLocalPressureV3HistoryMessageIDAllocationEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_v3_history_message_id_allocation_2026_07_02") {
		t.Fatal("missing v3 history message-id allocation measurement")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryMessageIDAfter) {
		t.Fatal("missing v3 history message-id allocation evidence ref")
	}
	before := loadPressureEvidence(t, localPressureV3HistoryAfter)
	after := loadPressureEvidence(t, localPressureV3HistoryMessageIDAfter)
	assertCleanPressureEvidence(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.08)
	assertAllocationReduced(t, before, after, "http_endpoint_threads_v3", 0.04)
}
