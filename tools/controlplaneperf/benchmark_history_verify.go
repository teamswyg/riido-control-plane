package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyBenchmarkHistory(root string, m manifest) error {
	if m.BenchmarkHistory == "" || m.BenchmarkHistoryCommand == "" {
		return fmt.Errorf("performance manifest must bind benchmark history")
	}
	if err := verifyBenchmarkHistoryCommand(m); err != nil {
		return err
	}
	return verifyBenchmarkHistoryFile(repoPath(root, m.BenchmarkHistory))
}

func verifyBenchmarkHistoryCommand(m manifest) error {
	command := m.BenchmarkHistoryCommand
	for _, needle := range []string{"-benchmark-in", "-append-benchmark-history", m.BenchmarkHistory} {
		if !strings.Contains(command, needle) {
			return fmt.Errorf("benchmark history command missing %q", needle)
		}
	}
	return nil
}

func verifyBenchmarkHistoryFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return scanBenchmarkHistory(file)
}
