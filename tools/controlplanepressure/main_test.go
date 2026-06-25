package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestControlPlanePressureProducesEvidence(t *testing.T) {
	out := t.TempDir() + "/pressure.json"
	err := mainRun([]string{
		"-duration", "20ms",
		"-concurrency", "1,2",
		"-threads", "3",
		"-lines", "2",
		"-evidence-out", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got pressureReport
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != evidenceSchema || len(got.Runs) != len(scenarios())*2 {
		t.Fatalf("report shape = %+v", got)
	}
	for _, run := range got.Runs {
		if run.Operations == 0 || run.Candidate.Next == "" {
			t.Fatalf("run missing evidence: %+v", run)
		}
		if run.ConcurrentUsers != run.Concurrency {
			t.Fatalf("concurrent_users = %d, want %d", run.ConcurrentUsers, run.Concurrency)
		}
		if run.Resources.CPUSecondsPerOp < 0 || run.Resources.CPUUtilizationPct < 0 {
			t.Fatalf("run has invalid CPU evidence: %+v", run.Resources)
		}
	}
}

func TestParseConcurrenciesRejectsInvalidValues(t *testing.T) {
	if _, err := parseConcurrencies("1,0"); err == nil {
		t.Fatal("expected invalid concurrency to fail")
	}
	if _, err := parseConfig([]string{"-duration", time.Second.String()}); err != nil {
		t.Fatal(err)
	}
}
