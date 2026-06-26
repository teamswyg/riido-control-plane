package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestControlPlanePressureWritesClosedLoopCandidates(t *testing.T) {
	root := t.TempDir()
	out := root + "/pressure.json"
	candidates := root + "/candidates.json"
	err := mainRun([]string{
		"-duration", "20ms", "-concurrency", "1", "-threads", "2", "-lines", "1",
		"-candidate-out", candidates, "-evidence-out", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got pressureCandidateEvidence
	data, err := os.ReadFile(candidates)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	assertPressureCandidateEvidence(t, got)
}

func assertPressureCandidateEvidence(t *testing.T, got pressureCandidateEvidence) {
	t.Helper()
	if got.SchemaVersion != pressureCandidateSchema || got.ID != pressureCandidateSourceID {
		t.Fatalf("candidate identity = %+v", got)
	}
	if got.CandidateCount != len(scenarios()) || len(got.Candidates) != len(scenarios()) {
		t.Fatalf("candidate count = %+v", got)
	}
	for _, candidate := range got.Candidates {
		if candidate.SourceRef.Run.ID == "" || candidate.PromotionEdge.Relation != "promotes_failure_to" {
			t.Fatalf("candidate missing closed-loop metadata: %+v", candidate)
		}
		if candidate.Measured.MaxConcurrentUsers == 0 || !candidate.Measured.ErrorFree {
			t.Fatalf("candidate missing pressure measurement: %+v", candidate)
		}
	}
}
