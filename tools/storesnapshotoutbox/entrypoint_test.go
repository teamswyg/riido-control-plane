package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMustRunAndMainSuccessPaths(t *testing.T) {
	out := filepath.Join(t.TempDir(), "must-run-evidence.json")
	mustRun([]string{"-repo", "../..", "-evidence-out", out})
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("mustRun evidence: %v", err)
	}

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	mainOut := filepath.Join(t.TempDir(), "main-evidence.json")
	os.Args = []string{"storesnapshotoutbox", "-repo", "../..", "-evidence-out", mainOut}
	main()
	if _, err := os.Stat(mainOut); err != nil {
		t.Fatalf("main evidence: %v", err)
	}
}
