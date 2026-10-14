package main

import (
	"encoding/json"
	"os"
	"testing"
)

const stagingBurst256Evidence = "docs/30-architecture/evidence/staging-public-burst-load-256-evidence-2026-07-02.json"

func TestStagingPublicBurst256EvidenceKeepsColdStartPartial(t *testing.T) {
	body, err := os.ReadFile("../../" + stagingBurst256Evidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted         bool           `json:"redacted"`
		AuthoritativeRun burst256Run    `json:"authoritative_run"`
		Findings         []burstFinding `json:"findings"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || evidence.AuthoritativeRun.Concurrency != 256 {
		t.Fatal("staging burst 256 evidence must be redacted and record concurrency 256")
	}
	if evidence.AuthoritativeRun.Failures != 0 ||
		evidence.AuthoritativeRun.LatencyMS.P95 > 250 ||
		evidence.AuthoritativeRun.LatencyMS.P99 > 650 {
		t.Fatal("staging burst 256 evidence exceeded warm-load guardrails")
	}
	if !hasFinding(evidence.Findings, "cold_start_not_exercised", "partial") {
		t.Fatal("staging burst 256 evidence must preserve cold-start partial finding")
	}
}

type burst256Run struct {
	Concurrency int `json:"concurrency"`
	Failures    int `json:"failures"`
	LatencyMS   struct {
		P95 int `json:"p95"`
		P99 int `json:"p99"`
	} `json:"latency_ms"`
}
