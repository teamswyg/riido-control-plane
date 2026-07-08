package main

import (
	"path/filepath"
	"testing"
)

func TestVerifySourceChecksRejectsMissingFileAndNeedle(t *testing.T) {
	root := t.TempDir()
	missing := []sourceCheck{{Name: "missing", File: "missing.go"}}
	if err := verifySourceChecks(root, missing); err == nil {
		t.Fatal("expected missing file error")
	}
	writeHealthReadyTestFile(t, filepath.Join(root, "source.go"), "needle")
	checks := []sourceCheck{{Name: "needle", File: "source.go", Contains: []string{"absent"}}}
	if err := verifySourceChecks(root, checks); err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestVerifyResultsRejectsMissingAndDrift(t *testing.T) {
	want := []endpointContract{
		{Name: "health", HTTPStatus: 200, Status: "ok"},
	}
	if err := verifyResults(want, nil); err == nil {
		t.Fatal("expected missing endpoint evidence")
	}
	got := []endpointEvidence{{Name: "health", HTTPStatus: 503, Status: "down"}}
	if err := verifyResults(want, got); err == nil {
		t.Fatal("expected endpoint drift error")
	}
}

func TestVerifyDocRejectsMissingAndStaleGeneratedDoc(t *testing.T) {
	root := t.TempDir()
	m := minimalHealthReadyManifest()
	if err := verifyDoc(root, m); err == nil {
		t.Fatal("expected missing generated doc error")
	}
	writeHealthReadyTestFile(t, filepath.Join(root, m.GeneratedDoc), "stale")
	if err := verifyDoc(root, m); err == nil {
		t.Fatal("expected stale generated doc error")
	}
}
