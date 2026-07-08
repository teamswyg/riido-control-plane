package main

import (
	"path/filepath"
	"testing"
)

func TestProviderStatusRunSurfacesRepoAndManifestErrors(t *testing.T) {
	if err := run(options{Repo: t.TempDir()}); err == nil {
		t.Fatal("expected missing repo root error")
	}
	root := t.TempDir()
	writeProviderStatusTestFile(t, filepath.Join(root, "go.mod"), "module test\n")
	if err := run(options{Repo: root, Manifest: "missing.json"}); err == nil {
		t.Fatal("expected missing manifest error")
	}
}

func TestProviderStatusRunSurfacesWriteDocAndEvidenceErrors(t *testing.T) {
	root := t.TempDir()
	writeProviderStatusTestFile(t, filepath.Join(root, "go.mod"), "module test\n")
	m := completeProviderStatusManifest()
	m.SourceChecks = []sourceCheck{{Name: "source", File: "source.go", Contains: []string{"needle"}}}
	writeProviderStatusTestFile(t, filepath.Join(root, "source.go"), "needle")
	parent := filepath.Join(root, "blocked")
	writeProviderStatusTestFile(t, parent, "not a dir")
	m.GeneratedDoc = filepath.Join("blocked", "doc.md")
	manifestPath := filepath.Join(root, "manifest.json")
	writeProviderStatusManifest(t, manifestPath, m)
	if err := run(options{Repo: root, Manifest: manifestPath, WriteDoc: true}); err == nil {
		t.Fatal("expected write doc error")
	}
	m.GeneratedDoc = "docs/generated.md"
	writeProviderStatusManifest(t, manifestPath, m)
	if err := run(options{Repo: root, Manifest: manifestPath, EvidenceOut: filepath.Join(parent, "evidence.json")}); err == nil {
		t.Fatal("expected evidence write error")
	}
}

func TestProviderStatusMainRunRejectsBadFlags(t *testing.T) {
	if err := mainRun([]string{"-unknown"}); err == nil {
		t.Fatal("expected flag error")
	}
}
