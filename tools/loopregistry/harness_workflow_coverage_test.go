package main

import "testing"

func TestHarnessWorkflowCoverageAcceptsManifest(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	if err := verifyHarnessWorkflowCoverage("../..", m); err != nil {
		t.Fatalf("harness workflow coverage: %v", err)
	}
}

func TestHarnessWorkflowCoverageRejectsUnregisteredWorkflow(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	m.HarnessWorkflowExclusions = nil
	if err := verifyHarnessWorkflowCoverage("../..", m); err == nil {
		t.Fatal("expected unregistered harness-like workflow to fail")
	}
}

func TestHarnessWorkflowCoverageRejectsStaleExclusion(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	m.HarnessWorkflowExclusions = append(m.HarnessWorkflowExclusions, harnessWorkflowExclusion{
		Workflow: ".github/workflows/missing-smoke.yml",
		Reason:   "stale exclusion must not pass",
	})
	if err := verifyHarnessWorkflowCoverage("../..", m); err == nil {
		t.Fatal("expected stale harness workflow exclusion to fail")
	}
}

func TestHarnessWorkflowCoverageRejectsReasonlessExclusion(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	m.HarnessWorkflowExclusions[0].Reason = ""
	if err := verifyHarnessWorkflowCoverage("../..", m); err == nil {
		t.Fatal("expected reasonless harness workflow exclusion to fail")
	}
}
