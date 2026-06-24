package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestCandidateFixtureIntakeVerifiesGeneratedCandidate(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/candidates.json"
	if err := promoteSummary(root, "docs/30-architecture/fixtures/harness-failure-summary.fixture.json", out); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), out); err != nil {
		t.Fatalf("verify candidate: %v", err)
	}
}

func promoteSummary(root, summary, out string) error {
	cmd := exec.Command("go", "run", "./tools/harnesspromotion",
		"-summary", summary, "-candidate-out", out)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "RIIDO_HARNESS_PROMOTION_NOW=2026-06-24T01:00:00Z")
	if body, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, body)
	}
	return nil
}
