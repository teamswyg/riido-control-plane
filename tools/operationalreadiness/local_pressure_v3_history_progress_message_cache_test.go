package main

import "testing"

const localPressureV3HistoryProgressMessageCacheAfter = "docs/30-architecture/evidence/control-plane-v3-history-progress-message-cache-after-2026-07-03.json"

func TestLocalPressureV3HistoryProgressMessageCacheEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_v3_history_progress_message_cache_2026_07_03") {
		t.Fatal("missing v3 history progress-message cache measurement")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryProgressMessageCacheAfter) {
		t.Fatal("missing v3 history progress-message cache evidence ref")
	}
	before := loadPressureEvidence(t, localPressureV3HistoryReadonlyProjectionAfter)
	after := loadPressureEvidence(t, localPressureV3HistoryProgressMessageCacheAfter)
	assertCleanPressureEvidence(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.10)
	assertCPUPerOpReduced(t, before, after, "thread_history_v3", 0.25)
}
