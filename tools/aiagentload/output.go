package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func writeReport(path string, r report) error {
	body, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if path == "" {
		_, err = os.Stdout.Write(body)
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func printSummary(r report) {
	fmt.Fprintf(os.Stderr,
		"scenario=%s total=%d success=%d failures=%d p95_ms=%d p99_ms=%d max_ms=%d\n",
		r.Scenario, r.Total, r.Success, r.Failures, r.Latency.P95, r.Latency.P99, r.Latency.Max,
	)
}
