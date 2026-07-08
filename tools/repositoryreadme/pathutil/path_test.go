package pathutil

import (
	"path/filepath"
	"testing"
)

func TestRepoPath(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "README.md")
	if got := RepoPath("/repo", abs); got != abs {
		t.Fatalf("absolute path = %q, want %q", got, abs)
	}
	if got := RepoPath("/repo", "docs/readme.md"); got != filepath.Join("/repo", "docs", "readme.md") {
		t.Fatalf("relative path = %q", got)
	}
}
