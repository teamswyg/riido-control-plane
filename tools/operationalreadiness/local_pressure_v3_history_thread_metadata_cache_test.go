package main

import "testing"

const localPressureV3HistoryThreadMetadataCacheAfter = "docs/30-architecture/evidence/control-plane-v3-history-thread-metadata-cache-after-2026-07-03.json"

func TestLocalPressureV3HistoryThreadMetadataCacheEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_v3_history_thread_metadata_cache_2026_07_03") {
		t.Fatal("missing v3 history thread metadata cache measurement")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryThreadMetadataCacheAfter) {
		t.Fatal("missing v3 history thread metadata cache evidence ref")
	}
	before := loadPressureEvidence(t, localPressureV3HistoryMessageCacheAfter)
	after := loadPressureEvidence(t, localPressureV3HistoryThreadMetadataCacheAfter)
	assertCleanPressureEvidence36(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.25)
	assertCPUPerOpReduced(t, before, after, "thread_history_v3", 0.25)
	assertAllocationReduced(t, before, after, "http_endpoint_threads_v3", 0.01)
	assertCPUPerOpReduced(t, before, after, "http_endpoint_threads_v3", 0.04)
}
