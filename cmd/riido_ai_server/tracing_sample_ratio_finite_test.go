package main

import "testing"

func TestTracingConfigRejectsNonFiniteSampleRatio(t *testing.T) {
	for _, value := range []string{"NaN", "+Inf", "-Inf"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(envTracingSampleRatio, value)
			if _, err := tracingConfigFromEnv(); err == nil {
				t.Fatal("non-finite sample ratio bypassed validation")
			}
		})
	}
}
