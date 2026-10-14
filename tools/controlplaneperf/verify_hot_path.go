package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func verifyHotPaths(root string, paths []hotPath) error {
	for _, path := range paths {
		if path.ID == "" || path.Category == "" || path.Risk == "" || path.Candidate == "" {
			return fmt.Errorf("hot path must bind id, category, risk, and candidate")
		}
		if len(path.Files) == 0 || len(path.Benchmarks)+len(path.Tests) == 0 {
			return fmt.Errorf("hot path %s must bind files and executable evidence", path.ID)
		}
		for _, file := range path.Files {
			if _, err := os.Stat(repoPath(root, file)); err != nil {
				return fmt.Errorf("hot path %s missing file %s: %w", path.ID, file, err)
			}
		}
		if err := verifyGoSymbols(root, path); err != nil {
			return err
		}
	}
	return nil
}

func verifyGoSymbols(root string, path hotPath) error {
	text, err := goTestText(root, path.Files)
	if err != nil {
		return err
	}
	for _, name := range path.Benchmarks {
		if !strings.Contains(text, "func "+name+"(") {
			return fmt.Errorf("hot path %s missing benchmark %s", path.ID, name)
		}
	}
	for _, name := range path.Tests {
		if !strings.Contains(text, "func "+name+"(") {
			return fmt.Errorf("hot path %s missing test %s", path.ID, name)
		}
	}
	return nil
}

func goTestText(root string, files []string) (string, error) {
	dir := filepath.Dir(files[0])
	matches, err := filepath.Glob(repoPath(root, filepath.Join(dir, "*_test.go")))
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			return "", err
		}
		b.Write(data)
	}
	return b.String(), nil
}
