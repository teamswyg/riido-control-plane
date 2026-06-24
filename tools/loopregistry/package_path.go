package main

import (
	"path/filepath"
	"strings"
)

func packagePathForFile(root, path string) (string, error) {
	rel, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return ".", nil
	}
	if strings.HasPrefix(rel, "../") {
		return rel, nil
	}
	return "./" + rel, nil
}
