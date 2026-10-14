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
	m := validPreCommitManifest()
	m.ID = ""
	if err := verifyManifest(m); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required field error, got %v", err)
	}
	m = validPreCommitManifest()
	m.EvidenceTTL = 0
	if err := verifyManifest(m); err == nil || !strings.Contains(err.Error(), "evidence_ttl") {
		t.Fatalf("expected ttl error, got %v", err)
	}
	m = validPreCommitManifest()
	m.Loop.Retrospective = ""
	if err := verifyManifest(m); err == nil || !strings.Contains(err.Error(), "retrospective") {
		t.Fatalf("expected loop retrospective error, got %v", err)
	}
}

func TestVerifySourcesRejectMissingInvalidAndIncompleteInputs(t *testing.T) {
	root := t.TempDir()
	m := validPreCommitManifest()
	if err := verifyPreCommitConfig(root, m, &verifyResult{}); err == nil {
		t.Fatal("expected missing pre-commit config error")
	}
	writeFile(t, filepath.Join(root, ".pre-commit-config.yaml"), "id: go-fmt\n")
	m.Hooks[0].ID = ""
	if err := verifyPreCommitConfig(root, m, &verifyResult{}); err == nil {
		t.Fatal("expected invalid hook error")
	}
	m = validPreCommitManifest()
	if err := verifyScripts(root, m, &verifyResult{}); err == nil {
		t.Fatal("expected missing script error")
	}
	writeFile(t, filepath.Join(root, "scripts/check.sh"), "go vet ./...\n")
	if err := verifyScripts(root, m, &verifyResult{}); err == nil {
		t.Fatal("expected missing script phrase error")
	}
}

func TestVerifyWorkflowAndDocRejectMissingOrStaleInputs(t *testing.T) {
	root := t.TempDir()
	m := validPreCommitManifest()
	if err := verifyWorkflow(root, m, &verifyResult{}); err == nil {
		t.Fatal("expected missing workflow error")
	}
	writeFile(t, filepath.Join(root, ".github/workflows/precommit.yml"), "on: push\n")
	if err := verifyWorkflow(root, m, &verifyResult{}); err == nil {
		t.Fatal("expected workflow schedule error")
	}
	if err := verifyDoc(root, m, "wanted"); err == nil {
		t.Fatal("expected missing doc error")
	}
	writeFile(t, filepath.Join(root, "docs/pre-commit.md"), "old")
	if err := verifyDoc(root, m, "wanted"); err == nil {
		t.Fatal("expected stale doc error")
	}
}
