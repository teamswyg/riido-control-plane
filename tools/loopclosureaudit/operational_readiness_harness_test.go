package main

import "testing"

func TestLoopClosureAuditCoversOperationalReadinessHarness(t *testing.T) {
	m, _ := loadForTest(t)
	req, ok := findRequirement(m.Requirements, "harness_failures_promote_to_closed_loop")
	if !ok {
		t.Fatal("harness promotion requirement is missing")
	}
	if !requirementHasLoop(req, "operational_readiness_release_harness") {
		t.Fatal("operational readiness harness loop proof is missing")
	}
	if !requirementHasGraphEdge(req, "operational_readiness_release_harness") {
		t.Fatal("operational readiness promotion edge proof is missing")
	}
}

func requirementHasLoop(req requirement, loopID string) bool {
	for _, check := range req.Checks {
		if check.Kind == "loop" && check.ID == loopID {
			return true
		}
	}
	return false
}

func requirementHasGraphEdge(req requirement, from string) bool {
	for _, check := range req.Checks {
		if check.Kind == "graph_edge" && check.From == from &&
			check.To == "closed_loop_candidate" {
			return true
		}
	}
	return false
}
