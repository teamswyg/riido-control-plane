package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGoCIBaselineBehaviorGolden(t *testing.T) {
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
	assertFileSHA256(t, out, "26fec9efc9d34f78e6cdd8d32c3c4c75c57aed75884899cbbc184dbc3512235b")
	doc := filepath.Join(repo, "docs/30-architecture/go-ci-baseline.md")
	assertFileSHA256(t, doc, "3067d72c4d2a04801711a939c9049e576eae5c350c8ee564b4b712eec1463bb8")
}

func assertFileSHA256(t *testing.T, path, want string) {
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
