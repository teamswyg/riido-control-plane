package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestRejectsMissingInvalidAndTrailingData(t *testing.T) {
	if _, err := loadManifest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatalf("expected missing manifest error")
	}
	root := t.TempDir()
	invalid := filepath.Join(root, "invalid.json")
	if err := os.WriteFile(invalid, []byte(`{"unknown":true}`), 0o644); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}
	if _, err := loadManifest(invalid); err == nil || !strings.Contains(err.Error(), "decode manifest") {
		t.Fatalf("expected decode error, got %v", err)
	}
	trailing := filepath.Join(root, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{} {}`), 0o644); err != nil {
		t.Fatalf("write trailing manifest: %v", err)
	}
	if _, err := loadManifest(trailing); err == nil || !strings.Contains(err.Error(), "one JSON object") {
		t.Fatalf("expected trailing object error, got %v", err)
	}
}

func TestRunWritesDocAndEvidenceOrReportsWriteFailure(t *testing.T) {
	root := tempRepo(t)
	manifestPath := writeManifest(t, root, baseManifest())
	out := filepath.Join(root, "out", "evidence.json")
	err := run(options{Repo: root, Manifest: manifestPath, EvidenceOut: out, WriteDoc: true})
	if err != nil {
		t.Fatalf("run writes doc/evidence: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "generated.md")); err != nil {
		t.Fatalf("generated doc missing: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("evidence missing: %v", err)
	}
	blocked := baseManifest()
	blocked.GeneratedDoc = "blocked/doc.md"
	if err := os.WriteFile(filepath.Join(root, "blocked"), []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}
	err = run(options{Repo: root, Manifest: writeManifest(t, root, blocked), WriteDoc: true})
	if err == nil || !strings.Contains(err.Error(), "write generated doc") {
		t.Fatalf("expected write generated doc error, got %v", err)
	}
}

func TestWriteJSONRejectsUnmarshalableValue(t *testing.T) {
	err := writeJSON(filepath.Join(t.TempDir(), "out.json"), make(chan int))
	if err == nil {
		t.Fatalf("expected json marshal error")
	}
}
