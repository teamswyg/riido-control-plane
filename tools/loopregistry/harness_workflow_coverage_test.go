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

func TestHarnessWorkflowCoverageRejectsUnboundReplacementClaim(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	m.HarnessWorkflowExclusions[0].ReplacementClaim = "missing_claim"
	if err := verifyHarnessWorkflowCoverage("../..", m); err == nil {
		t.Fatal("expected missing replacement claim to fail")
	}
}

func TestHarnessWorkflowCoverageRejectsReplacementClaimWithoutWorkflow(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	claim := claimsByID(m.Claims)[m.HarnessWorkflowExclusions[0].ReplacementClaim]
	claim.Files = nil
	for i := range m.Claims {
		if m.Claims[i].ID == claim.ID {
			m.Claims[i] = claim
		}
	}
	if err := verifyHarnessWorkflowCoverage("../..", m); err == nil {
		t.Fatal("expected workflow-unbound replacement claim to fail")
	}
}

func TestLoopRegistryEvidenceExposesHarnessWorkflowExclusions(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll(root, m, hashes)
	if err != nil {
		t.Fatalf("verifyAll: %v", err)
	}
	got := newEvidence(m, result, nil)
	if len(got.HarnessWorkflowExclusions) != len(m.HarnessWorkflowExclusions) {
		t.Fatalf("harness exclusions = %d, want %d", len(got.HarnessWorkflowExclusions), len(m.HarnessWorkflowExclusions))
	}
}
