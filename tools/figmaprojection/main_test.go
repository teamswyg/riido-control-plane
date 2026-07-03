package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "figma-projection.json")
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if got.SchemaVersion != evidenceSchema || got.Status != "verified" {
		t.Fatalf("unexpected evidence identity: %+v", got)
	}
	if got.ProjectionEntries != 16 || got.TotalAPIGeneratedAnnotations != 90 {
		t.Fatalf("unexpected evidence counts: %+v", got)
	}
}

func TestFigmaProjectionBehaviorGolden(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "figma-projection.json")
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
		t.Fatalf("run: %v", err)
	}
	evidenceBody, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	assertHash(t, "evidence", evidenceBody,
		"b1af55c4796f73aa9f91c0e94a5aec9cf90abaa06a48616f165cf75760c4c380")
	root, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	docBody, err := os.ReadFile(repoPath(root, defaultDoc))
	if err != nil {
		t.Fatalf("read generated doc: %v", err)
	}
	assertHash(t, "generated doc", docBody,
		"30daea6da8a963252840c01416d51283d57bf875e9401fe2a104b118e82e8469")
}

func assertHash(t *testing.T, label string, body []byte, want string) {
	t.Helper()
	got := fmt.Sprintf("%x", sha256.Sum256(body))
	if got != want {
		t.Fatalf("%s hash = %s", label, got)
	}
}
