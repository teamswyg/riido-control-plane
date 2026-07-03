package pathutil

import "path/filepath"

func Resolve(repo, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(repo, filepath.FromSlash(path))
}
