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

func TestRefreshWorkflowMustPublishLoopEvidence(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	m.Loops[0].Evidence[0].Path = "missing-loop-registry-evidence"
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected missing refresh artifact to fail")
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
