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
	report := agg.report(cfg, "staging.ai-api.riido.io", time.Unix(0, 0), time.Unix(1, 0), resourceDelta{}, pprofEvidence{})
	if report.Total != 1 || report.Success != 1 || report.BaseHost != "staging.ai-api.riido.io" {
		t.Fatalf("report = %+v", report)
	}
}

func TestAggregatorReportIncludesCapacityEvidence(t *testing.T) {
	agg := newAggregator()
	agg.add(result{Endpoint: "/healthz", Status: 200, Latency: 10 * time.Millisecond})
	agg.add(result{Endpoint: "/readyz", Status: 503, Latency: 50 * time.Millisecond})
	cfg := config{Scenario: "public", WorkspaceID: "workspace-a", Concurrency: 2}
	resources := resourceDelta{TotalAllocBytes: 1200, TotalAllocPerRequest: 600, Goroutines: 1}
	report := agg.report(cfg, "example.test", time.Unix(0, 0), time.Unix(2, 0), resources, pprofEvidence{})
	if report.RequestsPerSec != 1 || report.SuccessPerSec != 0.5 || report.FailureRatePct != 50 {
		t.Fatalf("rate evidence = rps %v success %v failure %v", report.RequestsPerSec, report.SuccessPerSec, report.FailureRatePct)
	}
	if report.Capacity.ConcurrentUsers != 2 || report.Capacity.P95LatencyMs != 10 || report.Capacity.ErrorFree {
		t.Fatalf("capacity evidence = %+v", report.Capacity)
	}
	if len(report.Findings) < 4 || report.Resource.TotalAllocPerRequest != 600 {
		t.Fatalf("findings/resource evidence = %d %+v", len(report.Findings), report.Resource)
	}
}
