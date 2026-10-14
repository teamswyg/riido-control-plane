package main

import "testing"

const localPressureProgressMessageIDStackAfter = "docs/30-architecture/evidence/control-plane-progress-message-id-stack-after-2026-07-03.json"

func TestLocalPressureProgressMessageIDStackEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_progress_message_id_stack_2026_07_03") {
		t.Fatal("missing progress message-id stack measurement")
	}
	if !hasEvidenceRef(check, localPressureProgressMessageIDStackAfter) {
		t.Fatal("missing progress message-id stack evidence ref")
	}
	before := loadPressureEvidence(t, localPressureV3HistoryPlainTextFastPathAfter)
	after := loadPressureEvidence(t, localPressureProgressMessageIDStackAfter)
	assertCleanPressureEvidence(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.15)
	assertAllocationReduced(t, before, after, "http_endpoint_threads_v3", 0.07)
}
