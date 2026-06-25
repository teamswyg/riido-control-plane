package main

import (
	"testing"
	"time"
)

func TestControlPlanePressureCoversV3ThreadHistoryEndpoint(t *testing.T) {
	if !hasScenario("http_endpoint_threads_v3") {
		t.Fatal("missing v3 thread history endpoint pressure scenario")
	}
	op, err := buildHTTPEndpointThreadsV3(config{
		Duration:      20 * time.Millisecond,
		Concurrencies: []int{1},
		Threads:       2,
		Lines:         1,
	})
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if err := op(); err != nil {
			t.Fatal(err)
		}
	}
}

func hasScenario(name string) bool {
	for _, scenario := range scenarios() {
		if scenario.name == name {
			return true
		}
	}
	return false
}
