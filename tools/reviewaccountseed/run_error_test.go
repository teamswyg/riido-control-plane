package main

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/reviewaccountseed/runconfig"
)

func TestMainRunRejectsBadFlag(t *testing.T) {
	if err := mainRun([]string{"-bad"}); err == nil {
		t.Fatal("expected flag parse error")
	}
}

func TestRunRejectsMissingRepoRootAndMalformedManifest(t *testing.T) {
	if err := run(runconfig.Options{Repo: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatal("expected repo root error")
	}
	root := t.TempDir()
	writeReviewSeedTestFile(t, filepath.Join(root, "go.mod"), "module test\n")
	manifestPath := filepath.Join(root, "manifest.json")
	writeReviewSeedTestFile(t, manifestPath, "{")
	if err := run(runconfig.Options{Repo: root, Manifest: manifestPath}); err == nil {
		t.Fatal("expected manifest error")
	}
}

func TestRunRejectsWriteDocCheckDocAndEvidenceFailures(t *testing.T) {
	m := minimalReviewSeedManifest()
	root, manifestPath := newReviewSeedRepo(t, m)
	writeReviewSeedTestFile(t, filepath.Join(root, "blocked"), "not a dir")
	err := run(runconfig.Options{Repo: root, Manifest: manifestPath, WriteDoc: true})
	if err != nil {
		t.Fatal(err)
	}
	m.GeneratedDoc = filepath.Join("blocked", "doc.md")
	root, manifestPath = newReviewSeedRepo(t, m)
	writeReviewSeedTestFile(t, filepath.Join(root, "blocked"), "not a dir")
	err = run(runconfig.Options{Repo: root, Manifest: manifestPath, WriteDoc: true})
	if err == nil {
		t.Fatal("expected generated doc write error")
	}
	err = run(runconfig.Options{Repo: root, Manifest: manifestPath, CheckDoc: true})
	if err == nil {
		t.Fatal("expected check doc read error")
	}
	err = run(runconfig.Options{
		Repo: root, Manifest: manifestPath, EvidenceOut: filepath.Join("blocked", "out.json"),
	})
	if err == nil {
		t.Fatal("expected evidence write error")
	}
}
