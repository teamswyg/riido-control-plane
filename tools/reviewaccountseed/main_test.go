package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/reviewaccountseed/requirements"
)

func TestReviewAccountSeedBehaviorGolden(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != requirements.EvidenceSchema || got.ID != "control-plane-review-account-seed" || got.Status != "verified" {
		t.Fatalf("unexpected evidence identity: %+v", got)
	}
	if got.CasesVerified != 4 || got.SourceChecks != 6 || got.SeedSSOT != "internal/riidoaiserver/review_account_seed.riido.json" {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.EvidenceArtifact != "review-account-seed-evidence" || got.Workflow != ".github/workflows/review-account-seed.yml" {
		t.Fatalf("unexpected evidence workflow: %+v", got)
	}
	assertReviewAccountSeedResults(t, got.Results)
	if got.Loop.Observation == "" || got.Loop.Retrospective == "" {
		t.Fatalf("missing loop evidence: %+v", got.Loop)
	}
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run reviewaccountseed: %v\n%s", err, body)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing evidence from go run: %v", err)
	}
}

func TestGeneratedDocMatchesManifest(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
