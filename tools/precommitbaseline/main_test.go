package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPreCommitBaselineBehaviorGolden(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	repo := filepath.Join("..", "..")
	out := filepath.Join(t.TempDir(), "evidence.json")
	err := run(options{
		Repo:        repo,
		Manifest:    defaultManifest,
		EvidenceOut: out,
		CheckDoc:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertPreCommitBaselineSHA256(t, out, "9fd9b7f1f793d909c0bd28eb71ca32bf47a609ffdfc0bd86494a75fb4e720964")
	doc := filepath.Join(repo, "docs/30-architecture/pre-commit-baseline.md")
	assertPreCommitBaselineSHA256(t, doc, "ea78ab5972d19c78f2e88e63ca5df83b2c4af00a24f6b228705167c37cdf583c")
}

func TestPreCommitBaselineManifest(t *testing.T) {
	err := run(options{
		Repo:        filepath.Join("..", ".."),
		Manifest:    defaultManifest,
		EvidenceOut: filepath.Join(t.TempDir(), "evidence.json"),
		CheckDoc:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPreCommitBaselineRejectsMissingPhrase(t *testing.T) {
	result := verifyResult{}
	err := requirePhrases("id: other", []string{"id: loop-registry-claim-binding"}, "hook", &result)
	if err == nil {
		t.Fatal("expected missing phrase failure")
	}
}

func assertPreCommitBaselineSHA256(t *testing.T, path, want string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := fmt.Sprintf("%x", sha256.Sum256(body))
	if got != want {
		t.Fatalf("%s sha256 = %s, want %s", path, got, want)
	}
}
