package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTracingSampleRatio = 0.01
	defaultTracingServiceName = "riido_ai_server"
	tracingShutdownTimeout    = 5 * time.Second
)

type tracingRuntimeConfig struct {
	Enabled      bool
	SampleRatio  float64
	OTLPEndpoint string
	ServiceName  string
}

type tracingShutdownFunc func(context.Context) error

type otelTraceRecorder struct {
	tracer trace.Tracer
}

type otelTraceSpan struct {
	span trace.Span
}

func tracingConfigFromEnv() (tracingRuntimeConfig, error) {
	enabled, err := envOptionalBool(envTracingEnabled)
	if err != nil {
		return tracingRuntimeConfig{}, err
	}
	sampleRatio, err := envOptionalFloat64(envTracingSampleRatio, defaultTracingSampleRatio)
	if err != nil {
		return tracingRuntimeConfig{}, err
	}
	if sampleRatio < 0 || sampleRatio > 1 {
		return tracingRuntimeConfig{}, fmt.Errorf("%s must be between 0 and 1", envTracingSampleRatio)
	}
	endpoint := strings.TrimSpace(os.Getenv(envTracingOTLPEndpoint))
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}
	serviceName := strings.TrimSpace(os.Getenv(envTracingServiceName))
	if serviceName == "" {
		serviceName = strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	}
	if serviceName == "" {
		serviceName = defaultTracingServiceName
	}
	return tracingRuntimeConfig{
		Enabled:      enabled,
		SampleRatio:  sampleRatio,
		OTLPEndpoint: endpoint,
		ServiceName:  serviceName,
	}, nil
}

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
	res, err := resource.New(ctx, resource.WithAttributes(attribute.String("service.name", config.ServiceName)))
	if err != nil {
		return nil, nil, fmt.Errorf("create trace resource: %w", err)
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.SampleRatio))),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return otelTraceRecorder{tracer: provider.Tracer(config.ServiceName)}, provider.Shutdown, nil
}

func shutdownTracing(shutdown tracingShutdownFunc) {
	if shutdown == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), tracingShutdownTimeout)
	defer cancel()
	_ = shutdown(ctx)
}

func otelTraceHTTPExporterOptions(endpoint string) []otlptracehttp.Option {
	parsed, err := url.Parse(endpoint)
	if err == nil && parsed.Scheme == "" {
		return []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(),
		}
	}
	options := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(endpoint)}
	if err == nil && parsed.Scheme == "http" {
		options = append(options, otlptracehttp.WithInsecure())
	}
	return options
}

//nolint:spancheck // Ownership is returned through otelTraceSpan and ended by the caller via the riidoaiserver.TraceSpan port.
func (r otelTraceRecorder) StartTraceSpan(ctx context.Context, start riidoaiserver.TraceSpanStart) (context.Context, riidoaiserver.TraceSpan) {
	if r.tracer == nil {
		return ctx, noopOtelTraceSpan{}
	}
	options := []trace.SpanStartOption{trace.WithSpanKind(otelSpanKind(start.Kind))}
	if len(start.Attributes) > 0 {
		options = append(options, trace.WithAttributes(otelAttributes(start.Attributes)...))
	}
	ctx, span := r.tracer.Start(ctx, start.Name, options...)
	return ctx, otelTraceSpan{span: span}
}

func (s otelTraceSpan) SetAttributes(attributes ...riidoaiserver.TraceAttribute) {
	if s.span == nil || len(attributes) == 0 {
		return
	}
	s.span.SetAttributes(otelAttributes(attributes)...)
}

func (s otelTraceSpan) RecordError(err error) {
	if s.span == nil || err == nil || errors.Is(err, context.Canceled) {
		return
	}
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

func (s otelTraceSpan) End() {
	if s.span != nil {
		s.span.End()
	}
}

type noopOtelTraceSpan struct{}

func (noopOtelTraceSpan) SetAttributes(...riidoaiserver.TraceAttribute) {}

func (noopOtelTraceSpan) RecordError(error) {}

func (noopOtelTraceSpan) End() {}

func otelSpanKind(kind riidoaiserver.TraceSpanKind) trace.SpanKind {
	switch kind {
	case riidoaiserver.TraceSpanKindServer:
		return trace.SpanKindServer
	case riidoaiserver.TraceSpanKindClient:
		return trace.SpanKindClient
	case riidoaiserver.TraceSpanKindInternal:
		return trace.SpanKindInternal
	default:
		return trace.SpanKindInternal
	}
}

func otelAttributes(attributes []riidoaiserver.TraceAttribute) []attribute.KeyValue {
	out := make([]attribute.KeyValue, 0, len(attributes))
	for _, attr := range attributes {
		key := strings.TrimSpace(attr.Key)
		if key == "" {
			continue
		}
		out = append(out, attribute.String(key, attr.Value))
	}
	return out
}

func envOptionalFloat64(key string, fallback float64) (float64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", key)
	}
	return value, nil
}
