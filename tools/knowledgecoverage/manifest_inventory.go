package main

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

func manifestInventory(root string) []string {
	var paths []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".riido.json") {
			paths = append(paths, slashPath(root, path))
		}
		return nil
	})
	sort.Strings(paths)
	return paths
}

func validateManifestInventory(root string, m manifest, docs []docClass) []string {
	tracked := trackedManifestSet(root, m, docs)
	var problems []string
	for _, path := range manifestInventory(root) {
		if !tracked[path] {
			problems = append(problems, "untracked executable manifest "+path)
		}
	}
	return problems
}

func manifestInventoryCount(root string) int {
	return len(manifestInventory(root))
}

func trackedManifestCount(root string, m manifest, docs []docClass) int {
	return len(trackedManifestSet(root, m, docs))
}

func untrackedManifests(root string, m manifest, docs []docClass) []string {
	tracked := trackedManifestSet(root, m, docs)
	var missing []string
	for _, path := range manifestInventory(root) {
		if !tracked[path] {
			missing = append(missing, path)
		}
	}
	return emptyIfNil(missing)
}
