package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyManifestRejectsMalformedManifest(t *testing.T) {
	if err := verifyManifest(manifest{}); err == nil || !strings.Contains(err.Error(), "schema_version") {
		t.Fatalf("expected schema error, got %v", err)
	}
	m := validManifest()
	m.ID = ""
	if err := verifyManifest(m); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required field error, got %v", err)
	}
	m = validManifest()
	m.Gates = nil
	if err := verifyManifest(m); err == nil || !strings.Contains(err.Error(), "gates") {
		t.Fatalf("expected gates error, got %v", err)
	}
	m = validManifest()
	m.Loop.Evaluate = ""
	if err := verifyManifest(m); err == nil || !strings.Contains(err.Error(), "evaluate") {
		t.Fatalf("expected loop evaluate error, got %v", err)
	}
}

func TestVerifyWorkflowRejectsMissingInvalidAndIncompleteWorkflow(t *testing.T) {
	root := t.TempDir()
	m := validManifest()
	if err := verifyWorkflow(root, m, &verifyResult{}); err == nil {
		t.Fatal("expected missing workflow error")
	}
	writeFile(t, filepath.Join(root, ".github/workflows/ci.yml"), "go test ./...\n")
	m.Gates[0].ID = ""
	if err := verifyWorkflow(root, m, &verifyResult{}); err == nil {
		t.Fatal("expected invalid gate error")
	}
	m = validManifest()
	m.Gates[0].Contains = []string{"go test ./...", "go vet ./..."}
	err := verifyWorkflow(root, m, &verifyResult{})
	if err == nil || !strings.Contains(err.Error(), "missing phrase") {
		t.Fatalf("expected missing phrase error, got %v", err)
	}
}

func TestVerifyDocRejectsMissingAndStaleDoc(t *testing.T) {
	root := t.TempDir()
	m := validManifest()
	if err := verifyDoc(root, m, "wanted"); err == nil {
		t.Fatal("expected missing doc error")
	}
	writeFile(t, filepath.Join(root, "docs/go-ci.md"), "old")
	err := verifyDoc(root, m, "wanted")
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected stale doc error, got %v", err)
	}
}
