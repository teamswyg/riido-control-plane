package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesAndChecksGeneratedDoc(t *testing.T) {
	repo, m := writeSmokeFixtureRepo(t)
	mustJSON(t, filepath.Join(repo, "manifest.json"), m)
	opt := options{Repo: repo, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true}
	if err := run(opt); err != nil {
		t.Fatal(err)
	}
	docPath := filepath.Join(repo, m.GeneratedDoc)
	if _, err := os.Stat(docPath); err != nil {
		t.Fatalf("missing generated doc: %v", err)
	}
	if err := os.WriteFile(docPath, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: repo, Manifest: "manifest.json", CheckDoc: true}); err == nil {
		t.Fatalf("expected stale generated doc error")
	}
}

func TestRunReportsWriteDocError(t *testing.T) {
	repo, m := writeSmokeFixtureRepo(t)
	blocker := filepath.Join(repo, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	m.GeneratedDoc = "blocker/doc.md"
	mustJSON(t, filepath.Join(repo, "manifest.json"), m)
	err := run(options{Repo: repo, Manifest: "manifest.json", WriteDoc: true})
	if err == nil {
		t.Fatalf("expected write generated doc error")
	}
}

func TestLoadSmokeMatrixRejectsMalformedInput(t *testing.T) {
	tmp := t.TempDir()
	if _, err := loadSmokeMatrix(filepath.Join(tmp, "missing.json")); err == nil {
		t.Fatalf("expected missing smoke matrix error")
	}
	path := filepath.Join(tmp, "smoke.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadSmokeMatrix(path); err == nil {
		t.Fatalf("expected invalid smoke matrix JSON error")
	}
}
