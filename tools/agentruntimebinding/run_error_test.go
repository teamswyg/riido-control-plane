package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesAndChecksGeneratedDoc(t *testing.T) {
	m := agentRuntimeBindingFixture()
	repo := writeAgentRuntimeBindingRepo(t, m)
	if err := run(options{Repo: repo, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true}); err != nil {
		t.Fatal(err)
	}
	writeAgentRuntimeBindingFile(t, filepath.Join(repo, m.GeneratedDoc), "stale")
	if err := run(options{Repo: repo, Manifest: "manifest.json", CheckDoc: true}); err == nil {
		t.Fatalf("expected stale generated doc error")
	}
}

func TestRunReportsFlagRootAndOutputErrors(t *testing.T) {
	if err := mainRun([]string{"-bad"}); err == nil {
		t.Fatalf("expected flag parse error")
	}
	if err := run(options{Repo: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatalf("expected missing repo root error")
	}
	m := agentRuntimeBindingFixture()
	repo := writeAgentRuntimeBindingRepo(t, m)
	blocker := filepath.Join(repo, "blocker")
	writeAgentRuntimeBindingFile(t, blocker, "file")
	err := run(options{Repo: repo, Manifest: "manifest.json", EvidenceOut: filepath.Join(blocker, "x.json")})
	if err == nil {
		t.Fatalf("expected write output error")
	}
}
