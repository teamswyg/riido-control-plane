package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	if err := run(options{Repo: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatal("run accepted missing repo")
	}
	if err := run(options{Repo: testRepoRoot(t), Manifest: "missing.json"}); err == nil {
		t.Fatal("run accepted missing manifest")
	}
}

func TestMainRunRejectsBadFlagsAndEvidenceWriteFailures(t *testing.T) {
	t.Parallel()
	if err := mainRun([]string{"-bad-flag"}); err == nil {
		t.Fatal("mainRun accepted bad flag")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "manifest.json")
	body := `{"schema_version":"x","id":"x","cases":[{"name":"bad","kind":"unknown"}]}`
	if err := os.WriteFile(manifestPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: root, Manifest: manifestPath}); err == nil {
		t.Fatal("run accepted unknown case kind")
	}
}

func TestRunWritesGeneratedDocForMinimalManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := `{"schema_version":"x","id":"x","title":"T","generated_doc":"doc.md"}`
	manifestPath := filepath.Join(root, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: root, Manifest: manifestPath, WriteDoc: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "doc.md")); err != nil {
		t.Fatalf("generated doc missing: %v", err)
	}
}
