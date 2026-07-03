package pathutil

import (
	"os"
	"path/filepath"
)

func FindRepoRoot(start string) (string, error) {
	root, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if exists(filepath.Join(root, "go.mod")) {
			return root, nil
		}
		next := filepath.Dir(root)
		if next == root {
			return "", os.ErrNotExist
		}
		root = next
	}
}

func Resolve(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
