package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesAndChecksGeneratedDoc(t *testing.T) {
	m := webFrontendAPIFixture()
	repo := writeWebFrontendRepo(t, m)
	opts := options{Repo: repo, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true}
	if err := run(opts); err != nil {
		t.Fatal(err)
	}
	writeWebFrontendFile(t, filepath.Join(repo, m.GeneratedDoc), "stale")
	if err := run(options{Repo: repo, Manifest: "manifest.json", CheckDoc: true}); err == nil {
		t.Fatalf("expected stale generated doc error")
	}
}

func TestRunReportsFlagRootRuntimeAndOutputErrors(t *testing.T) {
	if err := mainRun([]string{"-bad"}); err == nil {
		t.Fatalf("expected flag parse error")
	}
	if err := run(options{Repo: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatalf("expected missing repo root error")
	}
	m := webFrontendAPIFixture()
	m.CORSCases[0].WantAllowOrigin = "https://other.example"
	repo := writeWebFrontendRepo(t, m)
	if err := run(options{Repo: repo, Manifest: "manifest.json"}); err == nil {
		t.Fatalf("expected CORS verification error")
	}
	m = webFrontendAPIFixture()
	repo = writeWebFrontendRepo(t, m)
	blocker := filepath.Join(repo, "blocker")
	writeWebFrontendFile(t, blocker, "file")
	out := filepath.Join(blocker, "evidence.json")
	if err := run(options{Repo: repo, Manifest: "manifest.json", EvidenceOut: out}); err == nil {
		t.Fatalf("expected evidence write error")
	}
}
