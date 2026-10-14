package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyAllPropagatesScriptAndWorkflowFailures(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".pre-commit-config.yaml"), "id: go-fmt\n")
	m := validPreCommitManifest()
	if _, err := verifyAll(root, m); err == nil || !strings.Contains(err.Error(), "read script") {
		t.Fatalf("expected script read error, got %v", err)
	}
	writeFile(t, filepath.Join(root, "scripts/check.sh"), "go test ./...\n")
	writeFile(t, filepath.Join(root, ".github/workflows/precommit.yml"), "on: push\n")
	if _, err := verifyAll(root, m); err == nil || !strings.Contains(err.Error(), "schedule") {
		t.Fatalf("expected workflow schedule error, got %v", err)
	}
}
