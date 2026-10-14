package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func appendBenchmarkHistory(root string, m manifest, benchmarkIn, historyPath string) error {
	if benchmarkIn == "" {
		return fmt.Errorf("benchmark history append requires -benchmark-in")
	}
	if historyPath != m.BenchmarkHistory {
		return fmt.Errorf("benchmark history output must match manifest path")
	}
	text, err := readText(repoPath(root, benchmarkIn))
	if err != nil {
		return fmt.Errorf("read benchmark input: %w", err)
	}
	items, err := parseBenchmarkHistory(text)
	if err != nil {
		return err
	}
	record := newBenchmarkHistoryRecord(root, m, items)
	return appendBenchmarkHistoryRecord(repoPath(root, historyPath), record)
}

func appendBenchmarkHistoryRecord(path string, record benchmarkHistoryRecord) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return err
	}
	return nil
}
