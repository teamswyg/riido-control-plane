package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsManifestDecodeError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module test\n")
	writeFile(t, root, "manifest.json", "{")
	if err := run(options{Repo: root, Manifest: "manifest.json"}); err == nil {
		t.Fatal("expected manifest decode error")
	}
}

func TestRunReportsGeneratedDocWriteFailure(t *testing.T) {
	t.Parallel()
	m := testManifest()
	m.GeneratedDoc = "parent/out.md"
	root := writeTestRepo(t, m, "needle")
	if err := os.WriteFile(filepath.Join(root, "parent"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run(options{Repo: root, Manifest: "manifest.json", WriteDoc: true})
	if err == nil || !strings.Contains(err.Error(), "write generated doc") {
		t.Fatalf("run error = %v, want write generated doc", err)
	}
}

func TestRunReportsEvidenceWriteFailure(t *testing.T) {
	t.Parallel()
	root := writeTestRepo(t, testManifest(), "needle")
	if err := os.WriteFile(filepath.Join(root, "parent"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run(options{Repo: root, Manifest: "manifest.json", EvidenceOut: "parent/out.json"})
	if err == nil {
		t.Fatal("expected evidence write error")
	}
}
