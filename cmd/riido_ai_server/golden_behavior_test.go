package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

const runtimeConfigGoldenHash = "65a569793f3892f903809279798cc8895bd9b74b31698125db412b6db238d79a"

func TestRuntimeConfigBehaviorGolden(t *testing.T) {
	body := runtimeConfigGoldenBody(t)
	sum := sha256.Sum256(body)
	if got := hex.EncodeToString(sum[:]); got != runtimeConfigGoldenHash {
		t.Fatalf("runtime config golden drifted: %s\n%s", got, body)
	}
}

func runtimeConfigGoldenBody(t *testing.T) []byte {
	t.Helper()
	cases := []runtimeConfigSummary{
		runtimeConfigCase(t, map[string]string{}),
		runtimeConfigCase(t, map[string]string{
			envAddr:                   ":9090",
			envShutdownTimeoutSeconds: "7",
			envMetricsLogInterval:     "9",
			envPprofAddr:              "127.0.0.1:6060",
			envTracingEnabled:         "true",
			envTracingSampleRatio:     "0.5",
			envTracingOTLPEndpoint:    "https://otel.example.test",
			envTracingServiceName:     "riido-test",
			envWebAllowedOrigins:      "https://app.riido.io,http://localhost:5173/",
			envLongPollMaxHoldSeconds: "11",
			envLongPollTickSeconds:    "3",
		}),
	}
	body, err := json.MarshalIndent(cases, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func runtimeConfigCase(t *testing.T, env map[string]string) runtimeConfigSummary {
	t.Helper()
	clearRiidoAIServerEnv(t)
	for key, value := range env {
		t.Setenv(key, value)
	}
	config, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	return runtimeConfigSummary{
		Addr: config.Addr, ShutdownSeconds: int(config.ShutdownTimeout.Seconds()),
		MetricsSeconds: int(config.MetricsLogInterval.Seconds()), PprofAddr: config.PprofAddr,
		TracingEnabled: config.Tracing.Enabled, TracingSampleRatio: config.Tracing.SampleRatio,
		TracingEndpoint: config.Tracing.OTLPEndpoint, TracingService: config.Tracing.ServiceName,
		WebOrigins: config.WebAllowedOrigins, LongPollMaxHoldSeconds: int(config.LongPollMaxHold.Seconds()),
		LongPollTickSeconds: int(config.LongPollTick.Seconds()), AIAgentClientDev: config.AIAgentClientDev,
	}
}

type runtimeConfigSummary struct {
	Addr, PprofAddr, TracingEndpoint, TracingService string
	ShutdownSeconds, MetricsSeconds                  int
	TracingEnabled                                   bool
	TracingSampleRatio                               float64
	WebOrigins                                       []string
	LongPollMaxHoldSeconds, LongPollTickSeconds      int
	AIAgentClientDev                                 bool
}
