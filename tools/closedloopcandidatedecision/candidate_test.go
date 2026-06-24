package main

import (
	"os/exec"
	"testing"
)

func TestCandidateDecisionCoversGeneratedCandidate(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(root, out); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out); err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
}

func generateCandidate(root, out string) error {
	cmd := exec.Command("go", "run", "./tools/harnesspromotion",
		"-summary", "docs/30-architecture/fixtures/harness-failure-summary.fixture.json",
		"-candidate-out", out)
	cmd.Dir = root
	return cmd.Run()
}
