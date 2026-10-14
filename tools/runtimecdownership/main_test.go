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
	out := filepath.Join(t.TempDir(), "runtime-cd-ownership.json")
	if err := run([]string{"-check-doc", "-evidence-out", out}); err != nil {
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
	if got.Strategies != 3 || got.PublicPolicies < 4 || got.LoopFields != 5 {
		t.Fatalf("missing verification counts: %+v", got)
	}
}

func TestRuntimeCDOwnershipBehaviorGolden(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "runtime-cd-ownership.json")
	if err := run([]string{"-check-doc", "-evidence-out", out}); err != nil {
		t.Fatalf("run: %v", err)
	}
	evidenceBytes, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if got, want := sha256Hex(evidenceBytes), "87bd5bb3a44c14b55473ba66b564413095ce5a8c0b708cce0a80266f97365304"; got != want {
		t.Fatalf("evidence hash drifted: got %s want %s", got, want)
	}
	root, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}
	docBytes, err := os.ReadFile(repoPath(root, generatedDoc))
	if err != nil {
		t.Fatalf("read generated doc: %v", err)
	}
	if got, want := sha256Hex(docBytes), "8337012a7e235bb168f0666ed63fd266d6182cb8b5dbe15d61854841beebdabe"; got != want {
		t.Fatalf("generated doc hash drifted: got %s want %s", got, want)
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}
