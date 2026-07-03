package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindRepoRoot(start string) (string, error) {
	root, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
		next := filepath.Dir(root)
		if next == root {
			return "", fmt.Errorf("go.mod not found above %s", start)
		}
		root = next
	}
}

func RepoPath(repoRoot, slashPath string) string {
	if filepath.IsAbs(slashPath) {
		return slashPath
	}
	return filepath.Join(repoRoot, filepath.FromSlash(slashPath))
}
