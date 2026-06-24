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

func TestRefreshWorkflowCadenceMustFitExpiry(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	m.Loops[0].ExpiresAfterHours = 23
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected daily refresh to fail for 23h expiry")
	}
}

func loadLoopRegistryForTest(t *testing.T) (manifest, map[string]string) {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	hashes, err := claimHashes(root, m)
	if err != nil {
		t.Fatal(err)
	}
	return m, hashes
}

func findLoopIndexForTest(t *testing.T, m manifest, id string) int {
	t.Helper()
	for i, loop := range m.Loops {
		if loop.ID == id {
			return i
		}
	}
	t.Fatalf("loop %s not found", id)
	return -1
}
