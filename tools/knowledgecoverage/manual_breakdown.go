package main

import (
	"path"
	"slices"
	"strings"
)

func manualTopDirs(docs []docClass, limit int) []manualDir {
	counts := map[string]int{}
	for _, doc := range docs {
		if doc.Kind == "manual_registered" {
			counts[path.Dir(doc.Path)]++
		}
	}
	out := make([]manualDir, 0, len(counts))
	for dir, count := range counts {
		out = append(out, manualDir{Path: dir, Count: count})
	}
	slices.SortFunc(out, compareManualDir)
	if limit > 0 && len(out) > limit {
		return out[:limit]
	}
	return out
}

func compareManualDir(a, b manualDir) int {
	if a.Count != b.Count {
		return b.Count - a.Count
	}
	return strings.Compare(a.Path, b.Path)
}
