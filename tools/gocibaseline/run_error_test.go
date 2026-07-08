package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func writeBaselineManifest(t *testing.T, root string, m manifest) string {
	t.Helper()
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "baseline.json"), string(body))
	return "baseline.json"
}

func TestRunRejectsMissingRepoRootAndWriteDocFailure(t *testing.T) {
	missingRepo := filepath.Join(t.TempDir(), "missing")
	if err := run(options{Repo: missingRepo}); err == nil {
		t.Fatal("expected repo root error")
	}
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n")
	writeFile(t, filepath.Join(repo, ".github/workflows/ci.yml"), "go test ./...\n")
	m := validManifest()
	m.GeneratedDoc = "blocked/doc.md"
	writeFile(t, filepath.Join(repo, "blocked"), "file")
	err := run(options{Repo: repo, Manifest: writeBaselineManifest(t, repo, m), WriteDoc: true})
	if err == nil || !strings.Contains(err.Error(), "write generated doc") {
		t.Fatalf("expected write generated doc error, got %v", err)
	}
}
