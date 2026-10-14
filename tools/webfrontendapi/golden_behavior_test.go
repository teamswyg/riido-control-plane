package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestWebFrontendAPIBehaviorGolden(t *testing.T) {
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
	assertWebFrontendFileSHA256(t, out, "ad5a2105dc558ff6108dad73866b16cc762845c011d82c661da5ab23e804c629")
	doc := filepath.Join(repo, "docs/30-architecture/web-frontend-api.md")
	assertWebFrontendFileSHA256(t, doc, "f0abee974b574a266e30a0066981b6b54efaa22b83f68b7639df8bb65c1b5de1")
}

func assertWebFrontendFileSHA256(t *testing.T, path, want string) {
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
