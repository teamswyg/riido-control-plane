package main

import "testing"

func TestLoopClosureAuditRejectsMissingAssertions(t *testing.T) {
	m, deps := loadForTest(t)
	m.Assertions = nil
	if err := verifyAll("../..", m, deps); err == nil {
		t.Fatal("expected missing assertions to fail")
	}
}

func TestLoopClosureAuditEvidenceExposesAssertions(t *testing.T) {
	got := newEvidence(manifest{Assertions: []string{"proof must be actionable"}})
	if len(got.Assertions) != 1 {
		t.Fatalf("expected assertions in evidence, got %+v", got)
	}
}
