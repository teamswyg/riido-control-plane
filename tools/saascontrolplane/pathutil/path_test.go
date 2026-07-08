package pathutil

import (
	"path/filepath"
	"testing"
)

func TestRepoPath(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "contract.json")
	if got := RepoPath("/repo", abs); got != abs {
		t.Fatalf("absolute path = %q, want %q", got, abs)
	}
	if got := RepoPath("/repo", "docs/contract.json"); got != filepath.Join("/repo", "docs", "contract.json") {
		t.Fatalf("relative path = %q", got)
	}
}
