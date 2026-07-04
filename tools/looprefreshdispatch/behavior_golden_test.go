package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoopRefreshDispatchBehaviorGolden(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	t.Setenv("GITHUB_RUN_ID", "golden-run")
	t.Setenv("GITHUB_RUN_ATTEMPT", "2")
	t.Setenv("GITHUB_SHA", "0123456789abcdef")
	t.Setenv("GITHUB_REF_NAME", "main")
	t.Setenv("GITHUB_EVENT_NAME", "workflow_dispatch")
	dir := t.TempDir()
	root := repoRootForTest(t)
	dispatchOut := filepath.Join(dir, "dispatch.json")
	candidateOut := filepath.Join(dir, "candidates.json")
	args := []string{
		"-repo", root,
		"-commands-in", loopRefreshFixturePath(root),
		"-dispatch-out", dispatchOut,
		"-candidate-out", candidateOut,
	}
	if err := mainRun(args); err != nil {
		t.Fatal(err)
	}
	hash := hashLoopRefreshDispatchOutputs(t, dispatchOut, candidateOut)
	const expected = "ff8cc3383f023a4fd216864fb052278de8bce7c749f4a0c1f0c0321e4854b27f"
	if hash != expected {
		t.Fatalf("looprefreshdispatch golden sha = %s", hash)
	}
}

func hashLoopRefreshDispatchOutputs(t *testing.T, paths ...string) string {
	t.Helper()
	h := sha256.New()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		h.Write(data)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
