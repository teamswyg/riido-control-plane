package main

import "path/filepath"

func repoPath(repo, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(repo, filepath.FromSlash(path))
}

func stringSet(items []string) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, item := range items {
		out[item] = true
	}
	return out
}
