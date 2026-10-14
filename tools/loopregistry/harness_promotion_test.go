package main

import "testing"

func TestHarnessRefreshWorkflowMustRunPromotion(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	harness := loopIndex(t, m, "provider_acceptance_harness")
	m.Loops[harness].RefreshWorkflow = ".github/workflows/loop-registry.yml"
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected missing harnesspromotion command to fail")
	}
}

func TestHarnessMustPublishCandidateArtifact(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	harness := loopIndex(t, m, "provider_acceptance_harness")
	m.Loops[harness].Evidence[1].Path = "missing-closed-loop-candidates"
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected missing candidate artifact upload to fail")
	}
}

func loopIndex(t *testing.T, m manifest, id string) int {
	t.Helper()
	for idx, loop := range m.Loops {
		if loop.ID == id {
			return idx
		}
	}
	t.Fatalf("loop %s not found", id)
	return -1
}
