package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/requirements"
)

func TestRunWritesEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
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
	assertAIClientAPIGolden(t, got)
}

func assertAIClientAPIGolden(t *testing.T, got evidence) {
	t.Helper()
	if got.SchemaVersion != requirements.EvidenceSchema || got.ID != requirements.ExpectedID ||
		got.Status != "verified" || !got.SmokeMatrixParity || !got.GeneratedPathCovered {
		t.Fatalf("unexpected evidence identity/status: %+v", got)
	}
	if got.OperationCounts.Total != 57 || got.OperationCounts.V1 != 26 ||
		got.OperationCounts.V2 != 31 || got.OperationCounts.SmokeMatrix != 57 {
		t.Fatalf("operation counts drifted: %+v", got.OperationCounts)
	}
	if got.RequiredPaths != 12 || got.RuntimeConfigs != 1 || got.PublicFields != 24 ||
		got.DeploymentEvidence != 2 || got.ThreadHistoryV3Rules != 79 || got.SourceChecks != 39 {
		t.Fatalf("evidence counters drifted: %+v", got)
	}
	if got.Loop.Observation == "" || got.Loop.Hypothesis == "" || got.Loop.Execute == "" || got.Loop.Evaluate == "" ||
		got.Loop.Retrospective == "" {
		t.Fatalf("loop evidence must stay populated: %+v", got.Loop)
	}
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run aiagentclientapi: %v\n%s", err, body)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing evidence from go run: %v", err)
	}
}

func TestGeneratedDocMatchesManifest(t *testing.T) {
	if err := mainRun([]string{"-check-doc"}); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyAliasChecksGeneratedDoc(t *testing.T) {
	if err := mainRun([]string{"-verify"}); err != nil {
		t.Fatal(err)
	}
}
