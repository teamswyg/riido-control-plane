package main

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOtelTraceRecorderNilTracerReturnsNoopSpan(t *testing.T) {
	ctx := context.Background()
	gotCtx, span := (otelTraceRecorder{}).StartTraceSpan(ctx, riidoaiserver.TraceSpanStart{Name: "noop"})
	if gotCtx != ctx {
		t.Fatal("nil tracer changed context")
	}
	span.SetAttributes(riidoaiserver.StringTraceAttribute("ignored", "value"))
	span.RecordError(errors.New("ignored"))
	span.End()
	if _, ok := span.(noopOtelTraceSpan); !ok {
		t.Fatalf("span type = %T, want noopOtelTraceSpan", span)
	}
}

func TestOtelTraceNoopSpanMethodsAreSafe(t *testing.T) {
	span := noopOtelTraceSpan{}
	span.SetAttributes(riidoaiserver.StringTraceAttribute("ignored", "value"))
	span.RecordError(errors.New("ignored"))
	span.End()

	otelTraceSpan{}.SetAttributes(riidoaiserver.StringTraceAttribute("ignored", "value"))
	otelTraceSpan{}.SetAttributes()
}

func TestOtelTraceSpanIgnoresCanceledError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() { _ = provider.Shutdown(context.Background()) }()

	_, span := (otelTraceRecorder{tracer: provider.Tracer("riido-test")}).
		StartTraceSpan(context.Background(), riidoaiserver.TraceSpanStart{Name: "cancel"})
	span.RecordError(context.Canceled)
	span.End()

	spans := exporter.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("recorded spans = %d, want 1", len(spans))
	}
	if spans[0].Status().Code != codes.Unset {
		t.Fatalf("canceled status = %v, want unset", spans[0].Status().Code)
	}
}

func TestOtelTraceSpanIgnoresNilError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() { _ = provider.Shutdown(context.Background()) }()

	_, span := (otelTraceRecorder{tracer: provider.Tracer("riido-test")}).
		StartTraceSpan(context.Background(), riidoaiserver.TraceSpanStart{Name: "nil-error"})
	span.RecordError(nil)
	span.End()

	spans := exporter.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("recorded spans = %d, want 1", len(spans))
	}
	if spans[0].Status().Code != codes.Unset {
		t.Fatalf("nil error status = %v, want unset", spans[0].Status().Code)
	}
}
