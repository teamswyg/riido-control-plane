package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesBehaviorGoldenEvidence(t *testing.T) {
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
	assertGeneratedSmokeMatrixGolden(t, got)
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run aigeneratedsmokematrix: %v\n%s", err, body)
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

func TestWorkflowPublishesRefreshEvidence(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("../..", ".github/workflows/ai-agent-generated-endpoint-smoke-matrix.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"schedule:",
		"RIIDO_LOOP_IDS: closed_loop_candidate",
		"name: ai-agent-generated-endpoint-smoke-matrix-evidence",
		"if-no-files-found: error",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("workflow missing %q", want)
		}
	}
}
