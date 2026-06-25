package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCandidateDecisionCoversGeneratedCandidate(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out); err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
}

func TestCandidateDecisionRejectsExpiredCandidateArtifact(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-26T00:00:00Z")
	_, err = verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired candidate artifact failure, got %v", err)
	}
}

func generateCandidate(t *testing.T, root, out string) error {
	t.Helper()
	pinCandidateFreshnessClock(t)
	cmd := exec.Command("go", "run", "./tools/harnesspromotion",
		"-summary", "docs/30-architecture/fixtures/harness-failure-summary.fixture.json",
		"-candidate-out", out)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "RIIDO_HARNESS_PROMOTION_NOW=2026-06-24T01:00:00Z")
	if body, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, body)
	}
	return nil
}
