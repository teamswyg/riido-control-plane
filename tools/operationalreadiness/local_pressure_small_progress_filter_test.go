package main

import "testing"

const localPressureSmallProgressFilterAfter = "docs/30-architecture/evidence/control-plane-small-progress-filter-after-2026-07-03.json"

func TestLocalPressureSmallProgressFilterEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_small_progress_filter_2026_07_03") {
		t.Fatal("missing small progress filter measurement")
	}
	if !hasEvidenceRef(check, localPressureSmallProgressFilterAfter) {
		t.Fatal("missing small progress filter evidence ref")
	}
	before := loadPressureEvidence(t, localPressureV3HistoryThreadMetadataCacheAfter)
	after := loadPressureEvidence(t, localPressureSmallProgressFilterAfter)
	assertCleanPressureEvidence36(t, after)
	assertAllocationReduced(t, before, after, "progress_ingest_fragment", 0.40)
	assertCPUPerOpReduced(t, before, after, "progress_ingest_fragment", 0.15)
}
