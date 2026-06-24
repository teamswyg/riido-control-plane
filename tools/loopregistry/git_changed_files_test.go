package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitChangedFilesIncludesUntrackedFiles(t *testing.T) {
	root := t.TempDir()
	runGitForTest(t, root, "init")
	runGitForTest(t, root, "config", "user.email", "test@example.com")
	runGitForTest(t, root, "config", "user.name", "Test User")
	writeFileForTest(t, filepath.Join(root, "tracked.go"), "package main\n")
	runGitForTest(t, root, "add", "tracked.go")
	runGitForTest(t, root, "commit", "-m", "base")

	writeFileForTest(t, filepath.Join(root, "new_claim_surface.go"), "package main\n")
	files, err := gitChangedFiles(root, "HEAD")
	if err != nil {
		t.Fatalf("gitChangedFiles: %v", err)
	}
	if !files["new_claim_surface.go"] {
		t.Fatalf("untracked file missing from changed files: %+v", files)
	}
}

func runGitForTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFileForTest(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
