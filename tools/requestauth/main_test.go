package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const requestAuthGoldenSHA256 = "cf197cc59f6536f19a3f62bc5559073581930ee3a1aaf9791406655917883490"

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
	if got.SchemaVersion != evidenceSchema || got.Surfaces != len(requiredSurfaces) {
		t.Fatalf("unexpected evidence: %+v", got)
	}
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run requestauth: %v\n%s", err, body)
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

func TestRequestAuthorizationBehaviorGolden(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != requestAuthGoldenSHA256 {
		t.Fatalf("request authorization golden hash = %s", got)
	}
}
