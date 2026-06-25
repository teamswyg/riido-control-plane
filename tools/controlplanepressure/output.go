package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func writeReport(path string, report pressureReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if path == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func printSummary(report pressureReport) {
	for _, run := range report.Runs {
		fmt.Fprintf(os.Stderr, "scenario=%s concurrency=%d ops=%d errors=%d p95_us=%d alloc_per_op=%.1f goroutines_delta=%d\n",
			run.Scenario, run.Concurrency, run.Operations, run.Errors, run.LatencyUS.P95,
			run.Resources.TotalAllocPerOp, run.Resources.Goroutines)
	}
}
