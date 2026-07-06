package main

import (
	"strings"
	"testing"
)

func TestRuntimeTimingFromEnvRejectsInvalidDurations(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "shutdown timeout", key: envShutdownTimeoutSeconds},
		{name: "metrics interval", key: envMetricsLogInterval},
		{name: "assignment lease", key: envAssignmentActiveLease},
		{name: "long poll hold", key: envLongPollMaxHoldSeconds},
		{name: "long poll tick", key: envLongPollTickSeconds},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearRiidoAIServerEnv(t)
			t.Setenv(tt.key, "nope")
			_, err := runtimeTimingFromEnv()
			if err == nil || !strings.Contains(err.Error(), tt.key) {
				t.Fatalf("runtimeTimingFromEnv err=%v, want key %s", err, tt.key)
			}
		})
	}
}
