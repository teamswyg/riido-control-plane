package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyPrivateGatesRejectsInvalidAndDuplicateGates(t *testing.T) {
	if err := verifyPrivateGates([]privateGate{{Surface: "x"}}); err == nil {
		t.Fatalf("expected private gate field error")
	}
	gate := privateGate{Surface: "x", Owner: "owner", Evidence: "evidence"}
	if err := verifyPrivateGates([]privateGate{gate, gate}); err == nil {
		t.Fatalf("expected duplicate private gate error")
	}
}

func TestVerifyDocRejectsMissingAndStaleDoc(t *testing.T) {
	root := tempRepo(t)
	m := baseManifest()
	if err := verifyDoc(root, m, "wanted"); err == nil {
		t.Fatalf("expected missing generated doc error")
	}
	path := filepath.Join(root, "docs", "generated.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("write stale doc: %v", err)
	}
	if err := verifyDoc(root, m, "wanted"); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected stale doc error, got %v", err)
	}
}

func TestMainRunRejectsBadFlagsAndMissingRepo(t *testing.T) {
	if err := mainRun([]string{"-definitely-bad"}); err == nil {
		t.Fatalf("expected flag parse error")
	}
	missing := filepath.Join(t.TempDir(), "not-a-repo")
	if err := os.MkdirAll(missing, 0o755); err != nil {
		t.Fatalf("mkdir missing repo: %v", err)
	}
	if err := mainRun([]string{"-repo", missing}); err == nil {
		t.Fatalf("expected repo root error")
	}
}
