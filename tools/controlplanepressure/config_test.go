package main

import (
	"testing"
	"time"
)

func TestParseConcurrenciesRejectsInvalidValues(t *testing.T) {
	if _, err := parseConcurrencies("1,0"); err == nil {
		t.Fatal("expected invalid concurrency to fail")
	}
	cfg, err := parseConfig([]string{"-duration", time.Second.String(), "-scenario", "thread_history_v3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ScenarioNames) != 1 || cfg.ScenarioNames[0] != "thread_history_v3" {
		t.Fatalf("scenario filter = %+v", cfg.ScenarioNames)
	}
}
