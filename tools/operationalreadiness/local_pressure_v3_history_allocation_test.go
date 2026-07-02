package main

import "testing"

const localPressureV3HistoryAfter = "docs/30-architecture/evidence/control-plane-v3-history-allocation-after-2026-07-02.json"

func TestLocalPressureV3HistoryAllocationEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_v3_history_allocation_2026_07_02") {
		t.Fatal("missing v3 history allocation measurement")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryAfter) {
		t.Fatal("missing v3 history allocation evidence ref")
	}
	before := loadLocalPressureFollowup(t)
	after := loadPressureEvidence(t, localPressureV3HistoryAfter)
	assertCleanPressureEvidence(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.10)
	assertAllocationReduced(t, before, after, "http_endpoint_threads_v3", 0.05)
}

func assertCleanPressureEvidence(t *testing.T, evidence pressureEvidence) {
	t.Helper()
	if len(evidence.Runs) != 27 || len(evidence.Capacity) != 9 {
		t.Fatalf("unexpected pressure shape: runs=%d capacity=%d", len(evidence.Runs), len(evidence.Capacity))
	}
	for _, run := range evidence.Runs {
		if run.Errors != 0 || run.Resources.Goroutines != 0 {
			t.Fatalf("pressure run not clean: %+v", run)
		}
	}
}

func assertAllocationReduced(t *testing.T, before, after pressureEvidence, scenario string, minFraction float64) {
	t.Helper()
	beforeAlloc := capacityForScenario(t, before, scenario).AllocBytesPerOp
	afterAlloc := capacityForScenario(t, after, scenario).AllocBytesPerOp
	if afterAlloc >= beforeAlloc*(1-minFraction) {
		t.Fatalf("%s allocation not reduced enough: before=%f after=%f", scenario, beforeAlloc, afterAlloc)
	}
}

func capacityForScenario(t *testing.T, evidence pressureEvidence, scenario string) capacitySummary {
	t.Helper()
	for _, row := range evidence.Capacity {
		if row.Scenario == scenario {
			return row
		}
	}
	t.Fatalf("missing capacity row for %s", scenario)
	return capacitySummary{}
}
