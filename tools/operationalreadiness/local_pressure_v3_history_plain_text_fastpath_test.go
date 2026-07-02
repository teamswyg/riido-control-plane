package main

import "testing"

const localPressureV3HistoryPlainTextFastPathAfter = "docs/30-architecture/evidence/control-plane-v3-history-plain-text-fastpath-after-2026-07-03.json"

func TestLocalPressureV3HistoryPlainTextFastPathEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_v3_history_plain_text_fastpath_2026_07_03") {
		t.Fatal("missing v3 history plain-text fast-path measurement")
	}
	if !hasEvidenceRef(check, localPressureV3HistoryPlainTextFastPathAfter) {
		t.Fatal("missing v3 history plain-text fast-path evidence ref")
	}
	before := loadPressureEvidence(t, localPressureV3HistoryMessageIDAfter)
	after := loadPressureEvidence(t, localPressureV3HistoryPlainTextFastPathAfter)
	assertCleanPressureEvidence(t, after)
	assertAllocationReduced(t, before, after, "thread_history_v3", 0.30)
	assertAllocationReduced(t, before, after, "http_endpoint_threads_v3", 0.20)
	assertCPUPerOpReduced(t, before, after, "thread_history_v3", 0.25)
}

func assertCPUPerOpReduced(
	t *testing.T,
	before, after pressureEvidence,
	scenario string,
	minFraction float64,
) {
	t.Helper()
	beforeCPU := capacityForScenario(t, before, scenario).CPUSecondsPerOp
	afterCPU := capacityForScenario(t, after, scenario).CPUSecondsPerOp
	if afterCPU >= beforeCPU*(1-minFraction) {
		t.Fatalf("%s cpu/op not reduced enough: before=%f after=%f", scenario, beforeCPU, afterCPU)
	}
}
