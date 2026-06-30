package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseBenchmarkLine(line, pkg, goos, goarch string) (benchmarkHistoryItem, error) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return benchmarkHistoryItem{}, fmt.Errorf("invalid benchmark line: %s", line)
	}
	iterations, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return benchmarkHistoryItem{}, fmt.Errorf("invalid benchmark iterations %q", fields[1])
	}
	item := benchmarkHistoryItem{
		Name:       benchmarkName(fields[0]),
		Package:    pkg,
		GOOS:       goos,
		GOARCH:     goarch,
		Iterations: iterations,
	}
	for i := 2; i+1 < len(fields); i += 2 {
		value, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			continue
		}
		applyBenchmarkMetric(&item, fields[i+1], value)
	}
	if item.NsPerOp == 0 {
		return benchmarkHistoryItem{}, fmt.Errorf("benchmark line missing ns/op: %s", line)
	}
	return item, nil
}

func applyBenchmarkMetric(item *benchmarkHistoryItem, unit string, value float64) {
	switch unit {
	case "ns/op":
		item.NsPerOp = value
	case "B/op":
		item.BytesPerOp = value
	case "allocs/op":
		item.AllocsPerOp = value
	}
}
