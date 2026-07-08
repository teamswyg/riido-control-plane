package main

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/runconfig"
)

func TestRunWritesAndChecksGeneratedDoc(t *testing.T) {
	m := apiClientDeliveryFixture()
	repo := writeAPIClientDeliveryRepo(t, m)
	opt := runconfig.Options{Repo: repo, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true}
	if err := run(opt); err != nil {
		t.Fatal(err)
	}
	writeAPIClientFile(t, filepath.Join(repo, m.GeneratedDoc), "stale")
	if err := run(runconfig.Options{Repo: repo, Manifest: "manifest.json", CheckDoc: true}); err == nil {
		t.Fatalf("expected stale generated doc error")
	}
}

func TestRunReportsFlagRootAndOutputErrors(t *testing.T) {
	if err := mainRun([]string{"-bad"}); err == nil {
		t.Fatalf("expected flag parse error")
	}
	if err := run(runconfig.Options{Repo: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatalf("expected missing repo root error")
	}
	m := apiClientDeliveryFixture()
	repo := writeAPIClientDeliveryRepo(t, m)
	blocker := filepath.Join(repo, "blocker")
	writeAPIClientFile(t, blocker, "file")
	if err := run(runconfig.Options{Repo: repo, Manifest: "manifest.json", WriteDoc: true, EvidenceOut: filepath.Join(blocker, "x.json")}); err == nil {
		t.Fatalf("expected write output error")
	}
}
