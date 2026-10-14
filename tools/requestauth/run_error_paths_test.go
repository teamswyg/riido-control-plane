package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainRunRejectsBadFlag(t *testing.T) {
	t.Parallel()
	if err := mainRun([]string{"-nope"}); err == nil {
		t.Fatal("expected flag error")
	}
}

func TestRunWritesDocAndEvidenceInTempRepo(t *testing.T) {
	t.Parallel()
	root := writeTestRepo(t, testManifest(), "needle")
	opts := options{
		Repo: root, Manifest: "manifest.json", EvidenceOut: "out/evidence.json",
		WriteDoc: true, CheckDoc: true,
	}
	if err := run(opts); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"out/request.md", "out/evidence.json"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestRunFailsWhenGeneratedDocIsStale(t *testing.T) {
	t.Parallel()
	root := writeTestRepo(t, testManifest(), "needle")
	writeFile(t, root, "out/request.md", "stale")
	err := run(options{Repo: root, Manifest: "manifest.json", CheckDoc: true})
	if err == nil || !strings.Contains(err.Error(), "generated doc is stale") {
		t.Fatalf("run error = %v, want stale doc", err)
	}
}

func TestRunFailsWithoutRepoRoot(t *testing.T) {
	t.Parallel()
	err := run(options{Repo: t.TempDir(), Manifest: "manifest.json"})
	if err == nil || !strings.Contains(err.Error(), "go.mod") {
		t.Fatalf("run error = %v, want missing repo root", err)
	}
}
