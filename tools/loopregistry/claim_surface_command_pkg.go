package main

import "path/filepath"

func claimTestPackages(paths []string) map[string]bool {
	packages := map[string]bool{}
	for _, path := range paths {
		packages[packagePathFromRepoPath(path)] = true
	}
	delete(packages, "")
	return packages
}

func packagePathFromRepoPath(path string) string {
	dir := filepath.ToSlash(filepath.Dir(path))
	if dir == "." {
		return "."
	}
	return "./" + dir
}
