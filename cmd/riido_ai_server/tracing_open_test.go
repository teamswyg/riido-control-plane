package main

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestOpenTracingDisabledReturnsNil(t *testing.T) {
	recorder, shutdown, err := openTracing(context.Background(), tracingRuntimeConfig{})
	if err != nil {
		t.Fatalf("openTracing disabled: %v", err)
	}
	if recorder != nil || shutdown != nil {
		t.Fatalf("openTracing disabled returned recorder=%T shutdown=%v", recorder, shutdown != nil)
	}
}

func TestOpenTracingRequiresEndpointWhenEnabled(t *testing.T) {
	_, _, err := openTracing(context.Background(), tracingRuntimeConfig{
		Enabled: true, ServiceName: "riido-test", SampleRatio: 1,
	})
	if err == nil || !strings.Contains(err.Error(), envTracingOTLPEndpoint) {
		t.Fatalf("expected endpoint error, got %v", err)
	}
}

func TestOpenTracingCreatesRecorderAndShutdown(t *testing.T) {
	recorder, shutdown, err := openTracing(context.Background(), tracingRuntimeConfig{
		Enabled:      true,
		OTLPEndpoint: "http://127.0.0.1:4318",
		ServiceName:  "riido-test",
		SampleRatio:  1,
	})
	if err != nil {
		t.Fatalf("openTracing: %v", err)
	}
	defer otel.SetTracerProvider(noop.NewTracerProvider())
	if recorder == nil || shutdown == nil {
		t.Fatalf("openTracing recorder=%T shutdown=%v", recorder, shutdown != nil)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}
