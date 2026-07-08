package main

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/requirements"
)

func TestRunCoversWriteDocAndEvidencePaths(t *testing.T) {
	t.Parallel()
	m := validTestManifest()
	root := writeTestRepo(t, m, false)
	out := filepath.Join(root, "tmp/evidence.json")
	opts := options{Repo: root, Manifest: requirements.DefaultManifest, Boundary: "b00", WriteDoc: true, CheckDoc: true, EvidenceOut: out}
	if err := run(opts); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsIOAndSelectionFailures(t *testing.T) {
	t.Parallel()
	root := writeTestRepo(t, validTestManifest(), true)
	if err := run(options{Repo: root, Manifest: "missing.json"}); err == nil {
		t.Fatal("run accepted missing manifest")
	}
	if err := run(options{Repo: root, Manifest: requirements.DefaultManifest, Boundary: "missing"}); err == nil {
		t.Fatal("run accepted unknown boundary")
	}
	if err := mainRun([]string{"-definitely-not-a-flag"}); err == nil {
		t.Fatal("mainRun accepted unknown flag")
	}
}

func TestWriteHelpersRejectInvalidTargets(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFixtureFile(t, root, "blocked", "file")
	if err := writeText(filepath.Join(root, "blocked/doc.md"), "x"); err == nil {
		t.Fatal("writeText accepted file parent")
	}
	if err := writeJSON(filepath.Join(root, "ok.json"), func() {}); err == nil {
		t.Fatal("writeJSON accepted unsupported value")
	}
}
