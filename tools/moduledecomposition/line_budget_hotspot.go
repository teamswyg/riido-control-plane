package main

import (
	"path/filepath"
	"slices"
	"strings"
)

type lineBudgetHotspot struct {
	Path      string `json:"path"`
	Files     int    `json:"files"`
	MaxLines  int    `json:"max_lines"`
	TotalOver int    `json:"total_over_target_lines"`
}

func lineBudgetHotspots(samples []lineBudgetSample, target int) []lineBudgetHotspot {
	byDir := map[string]*lineBudgetHotspot{}
	for _, sample := range samples {
		dir := strings.TrimPrefix(filepath.ToSlash(filepath.Dir(sample.Path)), "./")
		hotspot := byDir[dir]
		if hotspot == nil {
			hotspot = &lineBudgetHotspot{Path: dir}
			byDir[dir] = hotspot
		}
		hotspot.Files++
		hotspot.TotalOver += sample.Lines - target
		if sample.Lines > hotspot.MaxLines {
			hotspot.MaxLines = sample.Lines
		}
	}
	return sortLineBudgetHotspots(byDir)
}

func sortLineBudgetHotspots(byDir map[string]*lineBudgetHotspot) []lineBudgetHotspot {
	out := make([]lineBudgetHotspot, 0, len(byDir))
	for _, hotspot := range byDir {
		out = append(out, *hotspot)
	}
	slices.SortFunc(out, compareLineBudgetHotspots)
	return out
}
