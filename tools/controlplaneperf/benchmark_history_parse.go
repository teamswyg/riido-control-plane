package main

import (
	"fmt"
	"strings"
)

func parseBenchmarkHistory(text string) ([]benchmarkHistoryItem, error) {
	var items []benchmarkHistoryItem
	var pkg, goos, goarch string
	for _, line := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(line, "goos:"):
			goos = strings.TrimSpace(strings.TrimPrefix(line, "goos:"))
		case strings.HasPrefix(line, "goarch:"):
			goarch = strings.TrimSpace(strings.TrimPrefix(line, "goarch:"))
		case strings.HasPrefix(line, "pkg:"):
			pkg = strings.TrimSpace(strings.TrimPrefix(line, "pkg:"))
		case strings.HasPrefix(line, "Benchmark"):
			item, err := parseBenchmarkLine(line, pkg, goos, goarch)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("benchmark history input has no benchmark rows")
	}
	return items, nil
}
