package main

import "time"

const evidenceSchema = "riido-control-plane-local-pressure-evidence.v1"

type pressureReport struct {
	SchemaVersion string           `json:"schema_version"`
	StartedAt     time.Time        `json:"started_at"`
	EndedAt       time.Time        `json:"ended_at"`
	DurationMs    int64            `json:"duration_ms"`
	Fixture       fixtureSummary   `json:"fixture"`
	Runs          []pressureRun    `json:"runs"`
	Candidates    []candidateEntry `json:"candidates"`
}

type fixtureSummary struct {
	Threads int `json:"threads"`
	Lines   int `json:"lines_per_thread"`
}

type pressureRun struct {
	Scenario        string         `json:"scenario"`
	Concurrency     int            `json:"concurrency"`
	ConcurrentUsers int            `json:"concurrent_users"`
	Operations      int64          `json:"operations"`
	Errors          int64          `json:"errors"`
	OpsPerSec       float64        `json:"ops_per_sec"`
	Latency         latencySummary `json:"latency_ms"`
	LatencyUS       latencySummary `json:"latency_us"`
	Resources       resourceDelta  `json:"resource_delta"`
	Candidate       candidateEntry `json:"candidate"`
}

type latencySummary struct {
	Min int64 `json:"min"`
	P50 int64 `json:"p50"`
	P90 int64 `json:"p90"`
	P95 int64 `json:"p95"`
	P99 int64 `json:"p99"`
	Max int64 `json:"max"`
}

type resourceDelta struct {
	HeapAllocBytes      int64   `json:"heap_alloc_bytes"`
	TotalAllocBytes     int64   `json:"total_alloc_bytes"`
	TotalAllocPerOp     float64 `json:"total_alloc_bytes_per_op"`
	Mallocs             int64   `json:"mallocs"`
	Frees               int64   `json:"frees"`
	Goroutines          int     `json:"goroutines"`
	UserCPUSeconds      float64 `json:"user_cpu_seconds"`
	SystemCPUSeconds    float64 `json:"system_cpu_seconds"`
	GCCPUSeconds        float64 `json:"gc_cpu_seconds"`
	ScavengeCPUSeconds  float64 `json:"scavenge_cpu_seconds"`
	ActiveCPUSeconds    float64 `json:"active_cpu_seconds"`
	AvailableCPUSeconds float64 `json:"available_cpu_seconds"`
	CPUUtilizationPct   float64 `json:"cpu_utilization_pct"`
	CPUSecondsPerOp     float64 `json:"cpu_seconds_per_op"`
}

type candidateEntry struct {
	Scenario string `json:"scenario"`
	Risk     string `json:"risk"`
	Next     string `json:"next"`
}
