package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConfigFromEnvParsesMetricsLogInterval(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envMetricsLogInterval, "15")
	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.MetricsLogInterval != 15*time.Second {
		t.Fatalf("metrics interval = %s", config.MetricsLogInterval)
	}
}

func TestConfigFromEnvParsesPprofAddr(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envPprofAddr, "127.0.0.1:6060")
	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.PprofAddr != "127.0.0.1:6060" {
		t.Fatalf("pprof addr = %q", config.PprofAddr)
	}
}

func TestConfigFromEnvParsesTracing(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envTracingEnabled, "true")
	t.Setenv(envTracingSampleRatio, "0.05")
	t.Setenv(envTracingOTLPEndpoint, "http://127.0.0.1:4318")
	t.Setenv(envTracingServiceName, "riido-ai-server-development")
	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if !config.Tracing.Enabled || config.Tracing.SampleRatio != 0.05 || config.Tracing.OTLPEndpoint != "http://127.0.0.1:4318" || config.Tracing.ServiceName != "riido-ai-server-development" {
		t.Fatalf("tracing config = %+v", config.Tracing)
	}
}

func TestTracingConfigRejectsInvalidSampleRatio(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envTracingEnabled, "true")
	t.Setenv(envTracingSampleRatio, "2")
	if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), envTracingSampleRatio) {
		t.Fatalf("expected tracing sample ratio error, got %v", err)
	}
}

func TestPprofHandlerServesIndex(t *testing.T) {
	server := newPprofServer("127.0.0.1:0")
	if server == nil {
		t.Fatal("pprof server should be configured")
	}
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	resp := httptest.NewRecorder()
	server.Handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "profile") {
		t.Fatalf("pprof response status=%d body=%s", resp.Code, resp.Body.String())
	}
	if newPprofServer("") != nil {
		t.Fatal("empty pprof addr should disable pprof server")
	}
}
