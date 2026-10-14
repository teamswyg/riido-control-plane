package main

import "testing"

func TestLoopClosureAuditRejectsEvidenceGraphSummaryDrift(t *testing.T) {
	m, deps := loadForTest(t)
	deps.graph.Chains[0].Decision = ""
	if err := verifyAll("../..", m, deps); err == nil {
		t.Fatal("expected graph summary drift to fail")
	}
}

func TestLoopClosureAuditRejectsNextLoopSummaryDrift(t *testing.T) {
	m, deps := loadForTest(t)
	deps.graph.Chains[0].NextLoop = ""
	if err := verifyAll("../..", m, deps); err == nil {
		t.Fatal("expected next-loop summary drift to fail")
	}
}
