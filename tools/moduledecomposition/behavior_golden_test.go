package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const moduleDecompositionGoldenSHA256 = "d3e3e02709e2bee9733ab4ceb849db8237809e5e46b687fc7e1432f4acd6cfe6"

func TestModuleDecompositionBehaviorGolden(t *testing.T) {
	repo := t.TempDir()
	mustWrite(t, repo, "go.mod", "module example.com/fixture\n\ngo 1.21\n")
	mustWrite(t, repo, "cmd/app/main.go", "package main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n")
	mustWrite(t, repo, "internal/app/app.go", "package app\nfunc Value() string { return \"ok\" }\n")
	mustWrite(t, repo, "tools/check/check.go", "package check\nfunc Check() {}\n")
	mustWrite(t, repo, "manifest.json", moduleDecompositionFixtureManifest)
	out := filepath.Join(repo, "evidence.json")
	if err := mainRun([]string{"-repo", repo, "-manifest", "manifest.json", "-evidence-out", out}); err != nil {
		t.Fatalf("mainRun: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != moduleDecompositionGoldenSHA256 {
		t.Fatalf("evidence SHA mismatch: got %s want %s\n%s", got, moduleDecompositionGoldenSHA256, body)
	}
}

func mustWrite(t *testing.T, root, slashPath, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(slashPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", slashPath, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", slashPath, err)
	}
}
