package main

import "time"

type config struct {
	BaseURL     string
	Token       string
	WorkspaceID string
	Scenario    string
	Duration    time.Duration
	Concurrency int
	Timeout     time.Duration
	OutputPath  string
}

type endpoint struct {
	Method string
	Path   string
	Auth   bool
	Body   string
}

type result struct {
	Endpoint string
	Status   int
	Latency  time.Duration
	Error    string
}

type report struct {
	SchemaVersion string                     `json:"schema_version"`
	Scenario      string                     `json:"scenario"`
	BaseHost      string                     `json:"base_host"`
	WorkspaceID   string                     `json:"workspace_id,omitempty"`
	StartedAt     time.Time                  `json:"started_at"`
	EndedAt       time.Time                  `json:"ended_at"`
	DurationMs    int64                      `json:"duration_ms"`
	Concurrency   int                        `json:"concurrency"`
	Total         int                        `json:"total"`
	Success       int                        `json:"success"`
	Failures      int                        `json:"failures"`
	StatusCounts  map[string]int             `json:"status_counts"`
	ErrorCounts   map[string]int             `json:"error_counts,omitempty"`
	Latency       latencySummary             `json:"latency_ms"`
	Endpoints     map[string]endpointSummary `json:"endpoints"`
}

type latencySummary struct {
	Min int64 `json:"min"`
	P50 int64 `json:"p50"`
	P90 int64 `json:"p90"`
	P95 int64 `json:"p95"`
	P99 int64 `json:"p99"`
	Max int64 `json:"max"`
}

type endpointSummary struct {
	Total        int            `json:"total"`
	Success      int            `json:"success"`
	Failures     int            `json:"failures"`
	StatusCounts map[string]int `json:"status_counts"`
	ErrorCounts  map[string]int `json:"error_counts,omitempty"`
	Latency      latencySummary `json:"latency_ms"`
}
