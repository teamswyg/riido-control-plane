package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesAndChecksGeneratedDoc(t *testing.T) {
	m := assignmentJournalFixture()
	repo := writeAssignmentJournalRepo(t, m)
	if err := run(options{Repo: repo, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true}); err != nil {
		t.Fatal(err)
	}
	docPath := filepath.Join(repo, m.GeneratedDoc)
	writeAssignmentFile(t, docPath, "stale")
	if err := run(options{Repo: repo, Manifest: "manifest.json", CheckDoc: true}); err == nil {
		t.Fatalf("expected stale generated doc error")
	}
}

func TestRunReportsWriteDocAndEvidenceErrors(t *testing.T) {
	m := assignmentJournalFixture()
	repo := writeAssignmentJournalRepo(t, m)
	blocker := filepath.Join(repo, "blocker")
	writeAssignmentFile(t, blocker, "file")
	m.GeneratedDoc = "blocker/doc.md"
	writeAssignmentJSON(t, filepath.Join(repo, "manifest.json"), m)
	if err := run(options{Repo: repo, Manifest: "manifest.json", WriteDoc: true}); err == nil {
		t.Fatalf("expected write generated doc error")
	}
	m = assignmentJournalFixture()
	writeAssignmentJSON(t, filepath.Join(repo, "manifest.json"), m)
	err := run(options{Repo: repo, Manifest: "manifest.json", EvidenceOut: filepath.Join(blocker, "x.json")})
	if err == nil {
		t.Fatalf("expected write evidence error")
	}
}

func TestMainRunRejectsBadFlagsAndMissingRoot(t *testing.T) {
	if err := mainRun([]string{"-bad"}); err == nil {
		t.Fatalf("expected flag parse error")
	}
	if err := run(options{Repo: filepath.Join(t.TempDir(), "missing"), Manifest: "manifest.json"}); err == nil {
		t.Fatalf("expected missing repo root error")
	}
}
