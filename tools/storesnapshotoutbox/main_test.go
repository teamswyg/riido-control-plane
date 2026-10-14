package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
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
	if got.Status != "verified" || got.CasesVerified != 3 || got.SourceChecks == 0 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	assertEvidenceHash(t, data)
}

func assertEvidenceHash(t *testing.T, data []byte) {
	t.Helper()
	sum := sha256.Sum256(data)
	want := "365891901acacfeb80988d775d5c6b48bd6549d74d0be621d3b7d919b702aecc"
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("evidence hash = %s, want %s", got, want)
	}
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run storesnapshotoutbox: %v\n%s", err, body)
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
