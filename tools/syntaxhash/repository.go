package main

import (
	"os"
	"path/filepath"
	"sort"
)

func scanRepository(root string, graph syntaxGraph, cfg constraints) (repoCoverage, error) {
	tracked := map[string]bool{}
	for _, target := range graph.Targets {
		for _, file := range target.FileHashes {
			tracked[file.Path] = true
		}
	}
	files, err := repositoryGoFiles(root)
	if err != nil {
		return repoCoverage{}, err
	}
	out := repoCoverage{GoFiles: len(files)}
	for _, file := range files {
		if tracked[file] {
			out.TrackedFiles++
			continue
		}
		if len(out.UntrackedSample) < cfg.UntrackedSampleLimit {
			out.UntrackedSample = append(out.UntrackedSample, file)
		}
	}
	out.UntrackedFiles = out.GoFiles - out.TrackedFiles
	out.CoveragePercent = percent(out.TrackedFiles, out.GoFiles)
	out.CoverageBasisPoints = basisPoints(out.TrackedFiles, out.GoFiles)
	return out, nil
}

func repositoryGoFiles(root string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(files)
	return files, err
}

func basisPoints(part, total int) int {
	if total == 0 {
		return 0
	}
	return part * 10000 / total
}
