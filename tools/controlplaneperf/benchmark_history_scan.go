package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func scanBenchmarkHistory(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	count := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var record benchmarkHistoryRecord
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&record); err != nil {
			return err
		}
		if err := verifyBenchmarkHistoryRecord(record); err != nil {
			return err
		}
		count++
	}
	if count == 0 {
		return fmt.Errorf("benchmark history has no records")
	}
	return scanner.Err()
}
