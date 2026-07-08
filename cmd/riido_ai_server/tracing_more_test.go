package main

import (
	"strings"
	"testing"
)

func TestTracingConfigAllowsSampleRatioBoundaries(t *testing.T) {
	for _, value := range []string{"0", "1"} {
		t.Run(value, func(t *testing.T) {
			clearRiidoAIServerEnv(t)
			t.Setenv(envTracingSampleRatio, value)
			config, err := tracingConfigFromEnv()
			if err != nil {
				t.Fatalf("tracingConfigFromEnv: %v", err)
			}
			if config.SampleRatio < 0 || config.SampleRatio > 1 {
				t.Fatalf("sample ratio = %f", config.SampleRatio)
			}
		})
	}
}

func TestEnvOptionalFloat64RejectsNonNumber(t *testing.T) {
	t.Setenv(envTracingSampleRatio, "not-a-number")
	if _, err := envOptionalFloat64(envTracingSampleRatio, 0.5); err == nil ||
		!strings.Contains(err.Error(), envTracingSampleRatio) {
		t.Fatalf("envOptionalFloat64 err=%v", err)
	}
}

func TestNoopOtelTraceSpanMethodsAreSafe(t *testing.T) {
	span := noopOtelTraceSpan{}
	span.SetAttributes()
	span.RecordError(nil)
	span.End()
}
