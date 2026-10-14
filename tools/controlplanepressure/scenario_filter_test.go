package main

import "testing"

func TestSelectedScenariosFiltersAndDeduplicates(t *testing.T) {
	cfg := config{ScenarioNames: []string{"thread_history_v3", "thread_history_v3"}}
	selected, err := selectedScenarios(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].name != "thread_history_v3" {
		t.Fatalf("selected scenarios = %+v", selected)
	}
}

func TestSelectedScenariosRejectsUnknownName(t *testing.T) {
	_, err := selectedScenarios(config{ScenarioNames: []string{"missing"}})
	if err == nil {
		t.Fatal("expected unknown scenario to fail")
	}
}
