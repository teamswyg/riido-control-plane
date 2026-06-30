package main

import (
	"os"
	"strings"
	"testing"
)

const benchmarkFixture = `goos: darwin
goarch: arm64
pkg: github.com/teamswyg/riido-control-plane/internal/riidoaiserver
BenchmarkHTTPTransactionMetricsObserve-10          12345  12.3 ns/op  4 B/op  1 allocs/op
BenchmarkRenderProgressMessage-10                  23456  45.6 ns/op  0 B/op  0 allocs/op
PASS
`

func TestParseBenchmarkHistory(t *testing.T) {
	items, err := parseBenchmarkHistory(benchmarkFixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items=%d", len(items))
	}
	if items[0].Name != "BenchmarkHTTPTransactionMetricsObserve" ||
		items[0].NsPerOp != 12.3 ||
		items[0].BytesPerOp != 4 ||
		items[0].AllocsPerOp != 1 {
		t.Fatalf("first item = %+v", items[0])
	}
}

func TestAppendBenchmarkHistoryRecord(t *testing.T) {
	path := t.TempDir() + "/history.jsonl"
	record := benchmarkHistoryRecord{
		SchemaVersion:    benchmarkHistorySchema,
		RecordedAt:       "2026-06-30T00:00:00Z",
		Source:           "go_test_benchmark",
		BenchmarkCommand: "go test -bench BenchmarkExample -benchmem",
		Benchmarks: []benchmarkHistoryItem{{
			Name: "BenchmarkExample", Iterations: 1, NsPerOp: 1,
		}},
	}
	if err := appendBenchmarkHistoryRecord(path, record); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), benchmarkHistorySchema) {
		t.Fatalf("history missing schema: %s", data)
	}
	if err := verifyBenchmarkHistoryFile(path); err != nil {
		t.Fatal(err)
	}
}

func TestBenchmarkHistoryRejectsUnknownFields(t *testing.T) {
	err := scanBenchmarkHistory(strings.NewReader(`{"schema_version":"` +
		benchmarkHistorySchema + `","recorded_at":"2026-06-30T00:00:00Z",` +
		`"source":"go_test_benchmark","benchmark_command":"go test",` +
		`"threshold_ms":1,"benchmarks":[{"name":"B","iterations":1,"ns_per_op":1}]}`))
	if err == nil {
		t.Fatal("expected unknown threshold field to fail")
	}
}
