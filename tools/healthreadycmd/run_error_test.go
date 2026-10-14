package main

import (
	"path/filepath"
	"testing"
)

func TestMainRunRejectsBadFlag(t *testing.T) {
	if err := mainRun([]string{"-unknown"}); err == nil {
		t.Fatal("expected flag parse error")
	}
}

func TestRunRejectsMissingRepoRootAndMalformedManifest(t *testing.T) {
	if err := run(options{Repo: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatal("expected repo root error")
	}
	root := t.TempDir()
	writeHealthReadyTestFile(t, filepath.Join(root, "go.mod"), "module test\n")
	manifestPath := filepath.Join(root, "manifest.json")
	writeHealthReadyTestFile(t, manifestPath, "{")
	if err := run(options{Repo: root, Manifest: manifestPath}); err == nil {
		t.Fatal("expected manifest error")
	}
}

func TestRunRejectsWriteDocAndEvidenceWriteFailures(t *testing.T) {
	m := minimalHealthReadyManifest()
	blocked := filepath.Join("blocked", "doc.md")
	m.GeneratedDoc = blocked
	root, manifestPath := newHealthReadyRepo(t, m)
	writeHealthReadyTestFile(t, filepath.Join(root, "blocked"), "not a dir")
	err := run(options{Repo: root, Manifest: manifestPath, WriteDoc: true})
	if err == nil {
		t.Fatal("expected generated doc write error")
	}
	root, manifestPath = newHealthReadyRepo(t, minimalHealthReadyManifest())
	writeHealthReadyTestFile(t, filepath.Join(root, "blocked"), "not a dir")
	err = run(options{
		Repo: root, Manifest: manifestPath, EvidenceOut: filepath.Join("blocked", "out.json"),
	})
	if err == nil {
		t.Fatal("expected evidence write error")
	}
}

func TestRunRejectsEndpointDriftThroughVerify(t *testing.T) {
	m := minimalHealthReadyManifest()
	m.Endpoints[0].Status = "down"
	root, manifestPath := newHealthReadyRepo(t, m)
	if err := run(options{Repo: root, Manifest: manifestPath}); err == nil {
		t.Fatal("expected endpoint verify error")
	}
}
