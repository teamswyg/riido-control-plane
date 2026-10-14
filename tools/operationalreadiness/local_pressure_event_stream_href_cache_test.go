package main

import "testing"

const localPressureEventStreamHrefCacheAfter = "docs/30-architecture/evidence/control-plane-event-stream-href-cache-after-2026-07-03.json"

func TestLocalPressureEventStreamHrefCacheEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_event_stream_href_cache_2026_07_03") {
		t.Fatal("missing event stream href cache measurement")
	}
	if !hasEvidenceRef(check, localPressureEventStreamHrefCacheAfter) {
		t.Fatal("missing event stream href cache evidence ref")
	}
	before := loadPressureEvidence(t, localPressureAgentSnapshotPointerMapAfter)
	after := loadPressureEvidence(t, localPressureEventStreamHrefCacheAfter)
	assertCleanPressureEvidence36(t, after)
	assertAllocationReduced(t, before, after, "thread_stream_subscription", 0.50)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.08)
}
