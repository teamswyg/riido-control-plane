package main

import (
	"path/filepath"
	"strings"
)

func manifestDocHasLoop(root, path string) bool {
	doc, ok := readJSONObject(root, path)
	if !ok {
		return false
	}
	loop, ok := doc["loop"].(map[string]any)
	if !ok {
		return false
	}
	for _, key := range []string{"observation", "hypothesis", "execute", "evaluate", "retrospective"} {
		value, ok := loop[key].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func manifestLoopSource(root, path string) (string, bool) {
	doc, ok := readJSONObject(root, path)
	if !ok {
		return "", false
	}
	source, ok := doc["loop_source"].(string)
	if !ok || strings.TrimSpace(source) == "" {
		return "", false
	}
	return manifestSourcePath(root, source)
}

func manifestSourcePath(root, source string) (string, bool) {
	path := filepath.Join(root, filepath.FromSlash(source))
	if filepath.IsAbs(source) {
		path = source
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(filepath.Clean(rel)), true
}
