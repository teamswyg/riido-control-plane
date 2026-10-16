package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type tracingShutdownFunc func(context.Context) error

func openTracing(ctx context.Context, config tracingRuntimeConfig) (riidoaiserver.TraceRecorder, tracingShutdownFunc, error) {
	if !config.Enabled {
		return nil, nil, nil
	}
	if strings.TrimSpace(config.OTLPEndpoint) == "" {
		return nil, nil, fmt.Errorf("%s or OTEL_EXPORTER_OTLP_ENDPOINT is required when %s is enabled", envTracingOTLPEndpoint, envTracingEnabled)
	}
	exporter, err := otlptracehttp.New(ctx, otelTraceHTTPExporterOptions(config.OTLPEndpoint)...)
	if err != nil {
		return nil, nil, fmt.Errorf("open OTLP trace exporter: %w", err)
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(tracingResource(config.ServiceName)),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.SampleRatio))),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetErrorHandler(otel.ErrorHandlerFunc(logOTelInternalError))
	return otelTraceRecorder{tracer: provider.Tracer(config.ServiceName)}, provider.Shutdown, nil
}

func logOTelInternalError(error) {
	log.Print("event=otel_internal_error")
}
