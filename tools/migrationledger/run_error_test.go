package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesAndChecksGeneratedDoc(t *testing.T) {
	m := migrationLedgerFixture()
	repo := writeMigrationLedgerRepo(t, m)
	if err := run(options{Repo: repo, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true}); err != nil {
		t.Fatal(err)
	}
	docPath := filepath.Join(repo, m.GeneratedDoc)
	writeMigrationFile(t, docPath, "stale")
	if err := run(options{Repo: repo, Manifest: "manifest.json", CheckDoc: true}); err == nil {
		t.Fatalf("expected stale generated doc error")
	}
}

func TestRunReportsWriteDocAndEvidenceErrors(t *testing.T) {
	m := migrationLedgerFixture()
	repo := writeMigrationLedgerRepo(t, m)
	blocker := filepath.Join(repo, "blocker")
	writeMigrationFile(t, blocker, "file")
	m.GeneratedDoc = "blocker/doc.md"
	writeMigrationJSON(t, filepath.Join(repo, "manifest.json"), m)
	if err := run(options{Repo: repo, Manifest: "manifest.json", WriteDoc: true}); err == nil {
		t.Fatalf("expected write generated doc error")
	}
	m = migrationLedgerFixture()
	writeMigrationJSON(t, filepath.Join(repo, "manifest.json"), m)
	err := run(options{Repo: repo, Manifest: "manifest.json", EvidenceOut: filepath.Join(blocker, "x.json")})
	if err == nil {
		t.Fatalf("expected write evidence error")
	}
}

func TestPathHelpersRejectMissingRootAndPreserveAbsolutePath(t *testing.T) {
	if _, err := findRepoRoot(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatalf("expected missing go.mod error")
	}
	abs := filepath.Join(t.TempDir(), "manifest.json")
	if got := repoPath("/repo", abs); got != abs {
		t.Fatalf("repoPath changed absolute path: %q", got)
	}
}
