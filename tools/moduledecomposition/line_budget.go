package main

import (
	"os"
	"path/filepath"
	"strings"
)

type lineBudgetResult struct {
	Target     int
	OverTarget int
	MaxLines   int
	Samples    []lineBudgetSample
	Hotspots   []lineBudgetHotspot
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
	result.Hotspots = topLineBudgetHotspots(samples, budget.HotspotLimit, budget.TargetLines)
	return result, nil
}
