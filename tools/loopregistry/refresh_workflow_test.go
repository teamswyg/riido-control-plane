package main

import "testing"

func TestExpiringLoopRequiresRefreshWorkflow(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	m.Loops[0].RefreshWorkflow = ""
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected missing refresh workflow to fail")
	}
}

func TestRefreshWorkflowMustBeScheduled(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	m.Loops[0].RefreshWorkflow = ".github/workflows/evidence-graph.yml"
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected unscheduled refresh workflow to fail")
	}
}

func TestLongExpiryLoopRequiresScheduledRefreshWorkflow(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	loop := findLoopIndexForTest(t, m, "open_decision_queue")
	m.Loops[loop].RefreshWorkflow = ".github/workflows/evidence-graph.yml"
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected long-expiry loop without scheduled refresh workflow to fail")
	}
}

func TestRefreshWorkflowMustPublishLoopEvidence(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	m.Loops[0].Evidence[0].Path = "missing-loop-registry-evidence"
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected missing refresh artifact to fail")
	}
}

func TestRefreshWorkflowMustPublishEveryLoopEvidenceArtifact(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	loop := findLoopIndexForTest(t, m, "closed_loop_candidate")
	for i := range m.Loops[loop].Evidence {
		m.Loops[loop].Evidence[i].RefreshWorkflow = ""
	}
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected missing artifact-specific refresh workflow to fail")
	}
}

func TestEvidenceSourceCanDeclareOwnRefreshWorkflow(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	loop := findLoopIndexForTest(t, m, "closed_loop_candidate")
	found := false
	for _, source := range m.Loops[loop].Evidence {
		if source.Path == "harness-promotion-evidence" &&
			source.RefreshWorkflow == ".github/workflows/harness-promotion.yml" {
			found = true
		}
	}
	if !found {
		t.Fatal("closed_loop_candidate must bind harness-promotion-evidence to its producer workflow")
	}
	if _, err := verifyAll("../..", m, hashes); err != nil {
		t.Fatalf("artifact-specific refresh workflow should verify: %v", err)
	}
}

func TestRefreshWorkflowCadenceMustFitExpiry(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	m.Loops[0].ExpiresAfterHours = 23
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected daily refresh to fail for 23h expiry")
	}
}
