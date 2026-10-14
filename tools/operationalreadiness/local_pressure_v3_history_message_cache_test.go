package main

import "testing"

const (
	localPressureV3HistoryMessageCacheBefore = "docs/30-architecture/evidence/control-plane-v3-history-message-cache-before-2026-07-03.json"
	localPressureV3HistoryMessageCacheAfter  = "docs/30-architecture/evidence/control-plane-v3-history-message-cache-after-2026-07-03.json"
)

func TestLocalPressureV3HistoryMessageCacheEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_v3_history_message_cache_2026_07_03") {
		t.Fatal("missing v3 history message cache measurement")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryMessageCacheBefore) {
		t.Fatal("missing v3 history message cache before evidence ref")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryMessageCacheAfter) {
		t.Fatal("missing v3 history message cache after evidence ref")
	}
	before := loadPressureEvidence(t, localPressureV3HistoryMessageCacheBefore)
	after := loadPressureEvidence(t, localPressureV3HistoryMessageCacheAfter)
	assertCleanPressureEvidence36(t, before)
	assertCleanPressureEvidence36(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.85)
	assertAllocationReduced(t, before, after, "http_endpoint_threads_v3", 0.25)
	assertCPUPerOpReduced(t, before, after, "thread_history_v3", 0.70)
}

func assertCleanPressureEvidence36(t *testing.T, evidence pressureEvidence) {
	t.Helper()
	if len(evidence.Runs) != 36 || len(evidence.Capacity) != 9 {
		t.Fatalf("unexpected pressure shape: runs=%d capacity=%d", len(evidence.Runs), len(evidence.Capacity))
	}
	for _, run := range evidence.Runs {
		if run.Errors != 0 || run.Resources.Goroutines != 0 {
			t.Fatalf("pressure run not clean: %+v", run)
		}
	}
}
