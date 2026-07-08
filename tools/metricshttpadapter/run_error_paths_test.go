package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/runconfig"
)

func TestRunWritesDocAndEvidenceInTempRepo(t *testing.T) {
	t.Parallel()
	root := writeTestRepo(t, testManifest(), "needle")
	opts := runconfig.Options{
		Repo: root, Manifest: "manifest.json", EvidenceOut: "out/evidence.json",
		WriteDoc: true, CheckDoc: true,
	}
	if err := run(opts); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"out/metrics.md", "out/evidence.json"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestRunFailsWhenGeneratedDocIsStale(t *testing.T) {
	t.Parallel()
	root := writeTestRepo(t, testManifest(), "needle")
	writeFile(t, root, "out/metrics.md", "stale")
	err := run(runconfig.Options{Repo: root, Manifest: "manifest.json", CheckDoc: true})
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("run error = %v, want stale doc", err)
	}
}

func TestRunFailsWithoutRepoRoot(t *testing.T) {
	t.Parallel()
	err := run(runconfig.Options{Repo: t.TempDir(), Manifest: "manifest.json"})
	if err == nil {
		t.Fatalf("run error = %v, want missing repo root", err)
	}
}
