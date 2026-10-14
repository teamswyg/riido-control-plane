package main

import "testing"

const (
	toolApprovalWaiterBefore = "docs/30-architecture/evidence/control-plane-tool-approval-waiter-signal-only-before-2026-07-03.json"
	toolApprovalWaiterAfter  = "docs/30-architecture/evidence/control-plane-tool-approval-waiter-signal-only-after-2026-07-03.json"
)

func TestLocalPressureToolApprovalSignalOnlyEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_tool_approval_signal_only_2026_07_03") {
		t.Fatal("missing tool approval signal-only measurement")
	}
	if !hasEvidenceRef(check, toolApprovalWaiterAfter) {
		t.Fatal("missing tool approval signal-only evidence ref")
	}
	before := loadPressureEvidence(t, toolApprovalWaiterBefore)
	after := loadPressureEvidence(t, toolApprovalWaiterAfter)
	assertCleanPressureEvidence(t, after)
	assertAllocationReduced(t, before, after, "tool_approval_waiters", 0.07)
	assertCPUPerOpReduced(t, before, after, "tool_approval_waiters", 0.30)
}
