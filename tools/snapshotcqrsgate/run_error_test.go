package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesAndChecksGeneratedDoc(t *testing.T) {
	m := snapshotGateFixture()
	repo := writeSnapshotGateRepo(t, m)
	if err := run([]string{"-repo", repo, "-manifest", "manifest.json", "-write-doc", "-check-doc"}); err != nil {
		t.Fatal(err)
	}
	writeSnapshotFile(t, filepath.Join(repo, m.GeneratedDoc), "stale")
	if err := run([]string{"-repo", repo, "-manifest", "manifest.json", "-check-doc"}); err == nil {
		t.Fatalf("expected stale generated doc error")
	}
}

func TestRunReportsFlagRootAndEvidenceErrors(t *testing.T) {
	if err := run([]string{"-bad"}); err == nil {
		t.Fatalf("expected flag parse error")
	}
	if err := run([]string{"-repo", filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatalf("expected missing repo root error")
	}
	m := snapshotGateFixture()
	repo := writeSnapshotGateRepo(t, m)
	blocker := filepath.Join(repo, "blocker")
	writeSnapshotFile(t, blocker, "file")
	err := run([]string{"-repo", repo, "-manifest", "manifest.json", "-evidence-out", filepath.Join(blocker, "x.json")})
	if err == nil {
		t.Fatalf("expected write evidence error")
	}
}
