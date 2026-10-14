package main

import (
	"strings"
	"testing"
	"time"
)

func TestEnvDurationSecondsRejectsNonPositiveValues(t *testing.T) {
	for _, value := range []string{"0", "-1", "nope"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(envShutdownTimeoutSeconds, value)
			if _, err := envDurationSeconds(envShutdownTimeoutSeconds, time.Second); err == nil || !strings.Contains(err.Error(), envShutdownTimeoutSeconds) {
				t.Fatalf("envDurationSeconds err=%v", err)
			}
		})
	}
}

func TestEnvOptionalDurationSecondsRejectsNonPositiveValues(t *testing.T) {
	for _, value := range []string{"0", "-1", "nope"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(envMetricsLogInterval, value)
			if _, err := envOptionalDurationSeconds(envMetricsLogInterval); err == nil || !strings.Contains(err.Error(), envMetricsLogInterval) {
				t.Fatalf("envOptionalDurationSeconds err=%v", err)
			}
		})
	}
}
