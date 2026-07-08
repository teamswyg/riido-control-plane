package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesGeneratedDoc(t *testing.T) {
	t.Parallel()
	repo := runtimeBoundaryTestRepo(t)
	docPath := "docs/runtime-boundary.md"
	manifestPath := "runtime-boundary.json"
	if err := writeJSON(repoPath(repo, manifestPath), runtimeBoundaryManifest(docPath)); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: repo, Manifest: manifestPath, WriteDoc: true}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(repoPath(repo, docPath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), generatedNotice) {
		t.Fatalf("generated doc missing notice: %s", body)
	}
}

func TestRunRejectsStaleGeneratedDoc(t *testing.T) {
	t.Parallel()
	repo := runtimeBoundaryTestRepo(t)
	manifestPath := "runtime-boundary.json"
	writeFile(t, repo, "docs/runtime-boundary.md", "stale")
	if err := writeJSON(repoPath(repo, manifestPath), runtimeBoundaryManifest("docs/runtime-boundary.md")); err != nil {
		t.Fatal(err)
	}
	err := run(options{Repo: repo, Manifest: manifestPath, CheckDoc: true})
	if err == nil || !strings.Contains(err.Error(), "is stale") {
		t.Fatalf("run stale doc err = %v", err)
	}
}

func TestLoadManifestRejectsTrailingObject(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, []byte(`{} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "one JSON object") {
		t.Fatalf("loadManifest trailing object err = %v", err)
	}
}
