package main

type pressureCandidateMeasurement struct {
	MaxConcurrentUsers int     `json:"max_concurrent_users"`
	OpsPerSec          float64 `json:"ops_per_sec"`
	P95LatencyUS       int64   `json:"p95_latency_us"`
	AllocBytesPerOp    float64 `json:"alloc_bytes_per_op"`
	CPUSecondsPerOp    float64 `json:"cpu_seconds_per_op"`
	GoroutineDelta     int     `json:"goroutine_delta"`
	ErrorFree          bool    `json:"error_free"`
}
