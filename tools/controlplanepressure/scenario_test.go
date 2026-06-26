package main

import "testing"

func TestScenariosIncludeSubscriberFanout(t *testing.T) {
	requireScenarios(t, "client_event_subscriber_fanout", "assignment_long_poll_wait", "tool_approval_waiters")
}

func requireScenarios(t *testing.T, names ...string) {
	t.Helper()
	seen := map[string]bool{}
	for _, scenario := range scenarios() {
		seen[scenario.name] = true
	}
	for _, name := range names {
		if !seen[name] {
			t.Fatalf("pressure evidence must keep scenario %s", name)
		}
	}
}
