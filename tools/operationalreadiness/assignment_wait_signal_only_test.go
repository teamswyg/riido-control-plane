package main

import "testing"

const (
	assignmentWaitSignalOnlyBefore = "docs/30-architecture/evidence/control-plane-assignment-wait-signal-only-before-2026-07-03.json"
	assignmentWaitSignalOnlyAfter  = "docs/30-architecture/evidence/control-plane-assignment-wait-signal-only-after-2026-07-03.json"
)

func TestLocalPressureAssignmentWaitSignalOnlyEvidence(t *testing.T) {
	check := readinessCheckByID(t, "testnet_load_capacity")
	if !hasMeasurement(check, "local_pressure_assignment_wait_signal_only_2026_07_03") {
		t.Fatal("missing assignment wait signal-only measurement")
	}
	if !hasEvidenceRef(check, assignmentWaitSignalOnlyAfter) {
		t.Fatal("missing assignment wait signal-only evidence ref")
	}
	before := loadPressureEvidence(t, assignmentWaitSignalOnlyBefore)
	after := loadPressureEvidence(t, assignmentWaitSignalOnlyAfter)
	assertCleanPressureEvidence(t, before)
	assertCleanPressureEvidence(t, after)
	assertAllocationReduced(t, before, after, "assignment_long_poll_wait", 0.20)
}
