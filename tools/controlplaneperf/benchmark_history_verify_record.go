package main

import "fmt"

func verifyBenchmarkHistoryRecord(record benchmarkHistoryRecord) error {
	if record.SchemaVersion != benchmarkHistorySchema || record.RecordedAt == "" {
		return fmt.Errorf("invalid benchmark history record identity")
	}
	if record.Source != "go_test_benchmark" || record.BenchmarkCommand == "" {
		return fmt.Errorf("invalid benchmark history source")
	}
	if len(record.Benchmarks) == 0 {
		return fmt.Errorf("benchmark history record has no benchmarks")
	}
	for _, item := range record.Benchmarks {
		if item.Name == "" || item.Iterations == 0 || item.NsPerOp == 0 {
			return fmt.Errorf("invalid benchmark history item %+v", item)
		}
	}
	return nil
}
