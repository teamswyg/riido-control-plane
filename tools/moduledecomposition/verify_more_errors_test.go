package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyPackageParityErrorPaths(t *testing.T) {
	t.Parallel()
	if err := verifyPackageParity(nil, []packageEntry{{Path: "cmd/app"}}); err == nil {
		t.Fatal("expected incomplete entry error")
	}
	entry := packageEntry{Path: "cmd/app", Kind: "runtime", Role: "entrypoint", MustNotOwn: "domain"}
	if err := verifyPackageParity(nil, []packageEntry{entry, entry}); err == nil {
		t.Fatal("expected duplicate package error")
	}
	if err := verifyPackageParity([]string{"cmd/app", "internal/app"}, []packageEntry{entry}); err == nil ||
		!strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected missing package error, got %v", err)
	}
	if err := verifyPackageParity([]string{"cmd/app"}, []packageEntry{
		entry,
		{Path: "tools/check", Kind: "tool", Role: "verification", MustNotOwn: "runtime"},
	}); err == nil || !strings.Contains(err.Error(), "count mismatch") {
		t.Fatalf("expected count mismatch error, got %v", err)
	}
}

func TestVerifyAllRejectsForbiddenImports(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/fixture\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	dir := filepath.Join(repo, "cmd", "app")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir package: %v", err)
	}
	source := []byte("package main\n\nimport _ \"net/http\"\n")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), source, 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	m := validShapeManifest()
	m.SourceRoots = []string{"cmd"}
	m.ForbiddenImports = []string{"net/http"}
	m.FileLineBudget.TargetLines = 100
	if _, err := verifyAll(repo, m); err == nil || !strings.Contains(err.Error(), "forbidden imports") {
		t.Fatalf("expected forbidden import error, got %v", err)
	}
}
