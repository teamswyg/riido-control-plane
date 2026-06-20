package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type lineBudgetResult struct {
	Target     int
	OverTarget int
	MaxLines   int
	Samples    []lineBudgetSample
}

type lineBudgetSample struct {
	Path  string `json:"path"`
	Lines int    `json:"lines"`
}

func scanLineBudget(repoRoot string, roots []string, budget fileLineBudget) (lineBudgetResult, error) {
	result := lineBudgetResult{Target: budget.TargetLines}
	if budget.TargetLines <= 0 {
		return result, nil
	}
	var samples []lineBudgetSample
	for _, root := range roots {
		err := filepath.WalkDir(repoPath(repoRoot, root), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry == nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			rel, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			lines, err := countLines(path)
			if err != nil {
				return err
			}
			if lines > result.MaxLines {
				result.MaxLines = lines
			}
			if lines > budget.TargetLines {
				result.OverTarget++
				samples = append(samples, lineBudgetSample{Path: filepath.ToSlash(rel), Lines: lines})
			}
			return nil
		})
		if err != nil {
			return result, err
		}
	}
	result.Samples = topLineBudgetSamples(samples, budget.SampleLimit)
	return result, nil
}

func countLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines, nil
}

func topLineBudgetSamples(samples []lineBudgetSample, limit int) []lineBudgetSample {
	slices.SortFunc(samples, func(a, b lineBudgetSample) int { return b.Lines - a.Lines })
	if limit > 0 && len(samples) > limit {
		return samples[:limit]
	}
	return samples
}
