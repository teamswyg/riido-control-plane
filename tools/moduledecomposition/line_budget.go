package main

import (
	"os"
	"path/filepath"
	"strings"
)

type lineBudgetResult struct {
	Target             int
	OverTarget         int
	MaxLines           int
	MaxFilesOverTarget int
	MaxFileLinesLimit  int
	Samples            []lineBudgetSample
	Hotspots           []lineBudgetHotspot
	HotspotRatchets    []lineBudgetHotspotRatchet
}

type lineBudgetSample struct {
	Path  string `json:"path"`
	Lines int    `json:"lines"`
}

func scanLineBudget(repoRoot string, roots []string, budget fileLineBudget) (lineBudgetResult, error) {
	result := lineBudgetResult{
		Target:             budget.TargetLines,
		MaxFilesOverTarget: budget.MaxFilesOverTarget,
		MaxFileLinesLimit:  budget.MaxFileLines,
	}
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
	hotspots := lineBudgetHotspots(samples, budget.TargetLines)
	result.Samples = topLineBudgetSamples(samples, budget.SampleLimit)
	result.Hotspots = trimLineBudgetHotspots(hotspots, budget.HotspotLimit)
	result.HotspotRatchets = buildLineBudgetHotspotRatchets(hotspots, budget.HotspotLimits)
	return result, nil
}
