package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesAndChecksGeneratedDoc(t *testing.T) {
	m := agentCatalogRBACFixture()
	repo := writeAgentCatalogRepo(t, m)
	opts := options{Repo: repo, Manifest: "manifest.json", Profile: "rbac", WriteDoc: true, CheckDoc: true}
	if err := run(opts); err != nil {
		t.Fatal(err)
	}
	writeAgentCatalogFile(t, filepath.Join(repo, m.GeneratedDoc), "stale")
	if err := run(options{Repo: repo, Manifest: "manifest.json", Profile: "rbac", CheckDoc: true}); err == nil {
		t.Fatalf("expected stale generated doc error")
	}
}

func TestRunReportsFlagRootProfileAndOutputErrors(t *testing.T) {
	if err := mainRun([]string{"-bad"}); err == nil {
		t.Fatalf("expected flag parse error")
	}
	if err := run(options{Repo: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatalf("expected missing repo root error")
	}
	m := agentCatalogRBACFixture()
	repo := writeAgentCatalogRepo(t, m)
	if err := run(options{Repo: repo, Manifest: "manifest.json", Profile: "missing"}); err == nil {
		t.Fatalf("expected unknown profile error")
	}
	blocker := filepath.Join(repo, "blocker")
	writeAgentCatalogFile(t, blocker, "file")
	opts := options{Repo: repo, Manifest: "manifest.json", Profile: "rbac", EvidenceOut: filepath.Join(blocker, "x.json")}
	if err := run(opts); err == nil {
		t.Fatalf("expected write output error")
	}
	m.GeneratedDoc = "blocker/generated.md"
	writeAgentCatalogJSON(t, filepath.Join(repo, "manifest.json"), m)
	if err := run(options{Repo: repo, Manifest: "manifest.json", Profile: "rbac", WriteDoc: true}); err == nil {
		t.Fatalf("expected write doc error")
	}
}
