package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitChangedFilesAndManifest(t *testing.T) {
	t.Parallel()
	root := newGitRepo(t)
	manifestPath := "docs/evidence.json"
	writeRepoFile(t, root, manifestPath, `{"schema_version":"v1","id":"base"}`)
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "base")
	writeRepoFile(t, root, "tracked.txt", "updated")
	writeRepoFile(t, root, "untracked.txt", "new")
	changed, err := gitChangedFiles(root, "HEAD")
	if err != nil {
		t.Fatalf("gitChangedFiles: %v", err)
	}
	if !changed["tracked.txt"] || !changed["untracked.txt"] {
		t.Fatalf("changed files = %+v", changed)
	}
	got, err := gitManifest(root, "HEAD", manifestPath)
	if err != nil {
		t.Fatalf("gitManifest: %v", err)
	}
	if got.ID != "base" {
		t.Fatalf("manifest id = %q", got.ID)
	}
}

func TestAddChangedFilesSkipsEmptyLines(t *testing.T) {
	t.Parallel()
	files := map[string]bool{}
	addChangedFiles(files, []byte("\nfirst\n\nsecond\n"))
	if len(files) != 2 || !files["first"] || !files["second"] {
		t.Fatalf("files = %+v", files)
	}
}

func newGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test")
	return root
}

func writeRepoFile(t *testing.T, root, path, body string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "core.hooksPath=/dev/null"}, args...)...)
	cmd.Dir = root
	cmd.Env = gitEnvWithoutOuterRepo(os.Environ())
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
