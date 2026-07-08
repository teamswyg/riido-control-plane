package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCoversWriteDocAndEvidencePaths(t *testing.T) {
	t.Parallel()
	m := validTestManifest()
	root := writeTestRepo(t, m, false)
	out := "tmp/context-map-evidence.json"
	opts := options{Repo: root, Manifest: defaultManifest, WriteDoc: true, CheckDoc: true, EvidenceOut: out}
	if err := run(opts); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(resolve(root, out)); err != nil {
		t.Fatalf("missing evidence: %v", err)
	}
}

func TestRunRejectsIOFailures(t *testing.T) {
	t.Parallel()
	if err := run(options{Repo: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatal("run accepted missing repo")
	}
	root := writeTestRepo(t, validTestManifest(), true)
	if err := run(options{Repo: root, Manifest: "missing.json"}); err == nil {
		t.Fatal("run accepted missing manifest")
	}
}

func TestRunRejectsWriteFailures(t *testing.T) {
	t.Parallel()
	m := validTestManifest()
	root := writeTestRepo(t, m, false)
	writeFixtureFile(t, root, "blocked", "file")
	m.GeneratedDoc = "blocked/doc.md"
	writeTestRepo(t, m, false)
	if err := writeText(resolve(root, m.GeneratedDoc), "x"); err == nil {
		t.Fatal("writeText accepted file parent")
	}
	if err := writeJSON(resolve(root, "blocked/evidence.json"), func() {}); err == nil {
		t.Fatal("writeJSON accepted unsupported value")
	}
}

func TestMainRunRejectsBadFlag(t *testing.T) {
	t.Parallel()
	if err := mainRun([]string{"-definitely-not-a-flag"}); err == nil {
		t.Fatal("mainRun accepted unknown flag")
	}
}
