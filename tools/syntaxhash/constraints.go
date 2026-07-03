package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func checkConstraints(root string, paths []string, cfg constraints) (constraintRun, error) {
	run := constraintRun{
		MaxFileLines:                    cfg.MaxFileLines,
		MaxFilesPerFolder:               cfg.MaxFilesPerFolder,
		MaxDepth:                        cfg.MaxDirectoryDepth,
		MinRepositoryCoverageBasisPoint: cfg.MinRepositoryCoverageBasisPoint,
	}
	for _, path := range paths {
		if err := walkConstraintRoot(root, path, cfg, &run); err != nil {
			return run, err
		}
	}
	if run.Violations > 0 {
		return run, fmt.Errorf("syntax hash artifact constraints failed")
	}
	return run, nil
}

func walkConstraintRoot(root, path string, cfg constraints, run *constraintRun) error {
	base := repoPath(root, path)
	counts := map[string]int{}
	return filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		dir := filepath.Dir(path)
		counts[dir]++
		if counts[dir] > cfg.MaxFilesPerFolder || depth(root, path) > cfg.MaxDirectoryDepth {
			run.Violations++
		}
		lines, err := lineCount(path)
		if err != nil {
			return err
		}
		if lines > cfg.MaxFileLines {
			run.Violations++
		}
		return nil
	})
}

func lineCount(path string) (int, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(body) == 0 {
		return 0, nil
	}
	return bytes.Count(body, []byte{'\n'}) + 1, nil
}

func depth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return 0
	}
	if rel == "." {
		return 0
	}
	return len(strings.Split(filepath.ToSlash(rel), "/"))
}
