package main

func newBenchmarkHistoryRecord(
	root string,
	m manifest,
	items []benchmarkHistoryItem,
) benchmarkHistoryRecord {
	git := readGitInfo(root)
	return benchmarkHistoryRecord{
		SchemaVersion:    benchmarkHistorySchema,
		RecordedAt:       formatEvidenceTime(evidenceNow()),
		GitCommit:        git.Commit,
		GitBranch:        git.Branch,
		GitDirty:         git.Dirty,
		Source:           "go_test_benchmark",
		BenchmarkCommand: m.BenchmarkCommand,
		Benchmarks:       append([]benchmarkHistoryItem(nil), items...),
	}
}
