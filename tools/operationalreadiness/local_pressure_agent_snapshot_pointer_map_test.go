package main

import "testing"

const localPressureAgentSnapshotPointerMapAfter = "docs/30-architecture/evidence/control-plane-agent-snapshot-pointer-map-after-2026-07-03.json"

func TestLocalPressureAgentSnapshotPointerMapEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_agent_snapshot_pointer_map_2026_07_03") {
		t.Fatal("missing agent snapshot pointer map measurement")
	}
	if !hasEvidenceRef(check, localPressureAgentSnapshotPointerMapAfter) {
		t.Fatal("missing agent snapshot pointer map evidence ref")
	}
	before := loadPressureEvidence(t, localPressureSmallProgressFilterAfter)
	after := loadPressureEvidence(t, localPressureAgentSnapshotPointerMapAfter)
	assertCleanPressureEvidence36(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.18)
	assertAllocationReduced(t, before, after, "http_endpoint_threads_v3", 0.01)
	assertCPUPerOpReduced(t, before, after, "thread_history_v3", 0.07)
}
