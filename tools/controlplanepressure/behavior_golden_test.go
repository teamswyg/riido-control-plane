package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

const controlPlanePressureGoldenSHA256 = "fe6365c6043db52791fe2872d775a4351c14098c779ca89f98891e6fb8350176"

func TestControlPlanePressureBehaviorGolden(t *testing.T) {
	t.Setenv("GITHUB_RUN_ID", "golden-run")
	t.Setenv("GITHUB_RUN_ATTEMPT", "1")
	t.Setenv("GITHUB_SHA", "0123456789abcdef")
	t.Setenv("GITHUB_REF_NAME", "golden")
	t.Setenv("GITHUB_EVENT_NAME", "push")
	sc := scenarios()[0]
	report := pressureReport{
		StartedAt: time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC),
		Candidates: []candidateEntry{
			candidateForScenario(sc),
		},
		Capacity: []capacityEstimate{{
			Scenario: sc.name, MaxConcurrentUsers: 8, OpsPerSec: 123.5,
			P95LatencyUS: 456, AllocBytesPerOp: 789.25, CPUSecondsPerOp: 0.000012,
			GoroutineDelta: 2, ErrorFree: true,
		}},
	}
	body, err := json.Marshal(pressureCandidateEvidenceFromReport(report))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != controlPlanePressureGoldenSHA256 {
		t.Fatalf("pressure candidate SHA mismatch: got %s want %s\n%s", got, controlPlanePressureGoldenSHA256, body)
	}
}
