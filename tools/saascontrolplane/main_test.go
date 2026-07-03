package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/requirements"
)

func TestSaaSControlPlaneBehaviorGolden(t *testing.T) {
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
	if got.SchemaVersion != requirements.EvidenceSchema || got.ID != requirements.ExpectedID || got.Status != "verified" {
		t.Fatalf("unexpected evidence identity: %+v", got)
	}
	if got.Boundaries != 19 || got.FocusedWorkflows != 17 || got.SourceChecks != 37 {
		t.Fatalf("unexpected coverage counts: %+v", got)
	}
	if got.SharedContracts != 3 || got.RequiredPhrases != 4 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run saascontrolplane: %v\n%s", err, body)
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
