package main

const benchmarkHistorySchema = "riido-control-plane-performance-benchmark-history.v1"

type benchmarkHistoryRecord struct {
	SchemaVersion    string                 `json:"schema_version"`
	RecordedAt       string                 `json:"recorded_at"`
	GitCommit        string                 `json:"git_commit"`
	GitBranch        string                 `json:"git_branch"`
	GitDirty         bool                   `json:"git_dirty"`
	Source           string                 `json:"source"`
	BenchmarkCommand string                 `json:"benchmark_command"`
	Benchmarks       []benchmarkHistoryItem `json:"benchmarks"`
}

type benchmarkHistoryItem struct {
	Name        string  `json:"name"`
	Package     string  `json:"package"`
	GOOS        string  `json:"goos,omitempty"`
	GOARCH      string  `json:"goarch,omitempty"`
	Iterations  int64   `json:"iterations"`
	NsPerOp     float64 `json:"ns_per_op"`
	BytesPerOp  float64 `json:"bytes_per_op,omitempty"`
	AllocsPerOp float64 `json:"allocs_per_op,omitempty"`
}
