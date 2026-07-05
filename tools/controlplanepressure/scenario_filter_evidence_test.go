package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestControlPlanePressureScenarioFilterProducesTargetedEvidence(t *testing.T) {
	out := t.TempDir() + "/pressure.json"
	err := mainRun([]string{
		"-duration", "20ms",
		"-concurrency", "1,2",
		"-scenario", "thread_history_v3",
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
	if len(got.Runs) != 2 || len(got.Candidates) != 1 || len(got.Capacity) != 1 {
		t.Fatalf("targeted report shape = %+v", got)
	}
	if got.Runs[0].Scenario != "thread_history_v3" || got.Runs[1].Scenario != "thread_history_v3" {
		t.Fatalf("unexpected scenarios = %+v", got.Runs)
	}
}
