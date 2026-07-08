package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	for _, name := range []string{"out/cloudwatch-emf.md", "out/evidence.json"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestRunFailsWhenGeneratedDocIsStale(t *testing.T) {
	t.Parallel()
	root := writeTestRepo(t, testManifest(), "needle")
	writeFile(t, root, "out/cloudwatch-emf.md", "stale")
	err := run(options{Repo: root, Manifest: "manifest.json", CheckDoc: true})
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("run error = %v, want stale doc", err)
	}
}

func TestRunFailsWithoutRepoRoot(t *testing.T) {
	t.Parallel()
	if err := run(options{Repo: t.TempDir(), Manifest: "manifest.json"}); err == nil {
		t.Fatal("expected missing repo root")
	}
}
