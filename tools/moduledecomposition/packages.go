package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func scanPackageDirs(repoRoot string, roots []string) ([]string, error) {
	found := map[string]bool{}
	for _, root := range roots {
		err := filepath.WalkDir(repoPath(repoRoot, root), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry == nil {
				return err
			}
			if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			if entry.IsDir() {
				return nil
			}
			if strings.HasSuffix(entry.Name(), ".go") {
				dir, err := filepath.Rel(repoRoot, filepath.Dir(path))
				if err != nil {
					return err
				}
				found[filepath.ToSlash(dir)] = true
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	packages := make([]string, 0, len(found))
	for path := range found {
		packages = append(packages, path)
	}
	slices.Sort(packages)
	return packages, nil
}
