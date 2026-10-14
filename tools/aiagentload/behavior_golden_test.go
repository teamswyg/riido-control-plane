package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

const aiAgentLoadGoldenSHA256 = "714378884a2dc2c6ecc0a8e1e1d62e7b36d7ac7ce4720f56ab1f26a32ba9b11a"

func TestAIAgentLoadBehaviorGolden(t *testing.T) {
	agg := newAggregator()
	agg.add(result{Endpoint: "/healthz", Status: 200, Latency: 10 * time.Millisecond})
	agg.add(result{Endpoint: "/readyz", Status: 503, Latency: 50 * time.Millisecond})
	agg.add(result{Endpoint: "/v2/client/workspaces/ws/ai-agent/devices", Status: 200, Latency: 25 * time.Millisecond})
	agg.add(result{Endpoint: "/v2/client/workspaces/ws/ai-agent/devices", Status: 200, Latency: 30 * time.Millisecond})
	cfg := config{Scenario: "client-read", Token: "secret-token", WorkspaceID: "ws", Concurrency: 3}
	resources := resourceDelta{
		HeapAllocBytes: 1024, TotalAllocBytes: 3200, TotalAllocPerRequest: 800,
		Mallocs: 30, Frees: 12, Goroutines: 1,
	}
	pprof := pprofEvidence{
		Enabled: true, BaseHost: "127.0.0.1:6060", ProfileSeconds: 1,
		Samples: []pprofSample{{Name: "heap", Path: "/debug/pprof/heap", Status: 200, Bytes: 128, LatencyMs: 3}},
	}
	report := agg.report(
		cfg, "staging.ai-api.riido.io",
		time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 24, 0, 0, 3, 0, time.UTC),
		resources, pprof,
	)
	body, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != aiAgentLoadGoldenSHA256 {
		t.Fatalf("load report SHA mismatch: got %s want %s\n%s", got, aiAgentLoadGoldenSHA256, body)
	}
}
