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
	assertFileSHA256(t, out, "b8c46b600f5557dd5feb57c47c4419bcfcd5e586fabb1d49203c3b909adfe535")
	doc := filepath.Join(repo, "docs/30-architecture/go-ci-baseline.md")
	assertFileSHA256(t, doc, "890a178a80a1a2b1508220b325af4ccb9db47f3efcb6e2cabba66163e2484df7")
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
