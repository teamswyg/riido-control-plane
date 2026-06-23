package main

import (
	"testing"
	"time"
)

func TestSummarizeLatency(t *testing.T) {
	values := []time.Duration{
		100 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		1000 * time.Millisecond,
	}
	got := summarizeLatency(values)
	if got.Min != 10 || got.P50 != 50 || got.P90 != 100 || got.P99 != 100 || got.Max != 1000 {
		t.Fatalf("latency summary = %+v", got)
	}
}

func TestAggregatorReportRedactsToken(t *testing.T) {
	agg := newAggregator()
	agg.add(result{Endpoint: "/healthz", Status: 200, Latency: time.Millisecond})
	cfg := config{Scenario: "client-read", Token: "secret-token", WorkspaceID: "workspace-a", Concurrency: 1}
	report := agg.report(cfg, "staging.ai-api.riido.io", time.Unix(0, 0), time.Unix(1, 0))
	if report.Total != 1 || report.Success != 1 || report.BaseHost != "staging.ai-api.riido.io" {
		t.Fatalf("report = %+v", report)
	}
}
