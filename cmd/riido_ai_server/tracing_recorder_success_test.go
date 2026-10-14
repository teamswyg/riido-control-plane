package main

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestOtelTraceRecorderRecordsSpanAttributesAndErrors(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() { _ = provider.Shutdown(context.Background()) }()

	recorder := otelTraceRecorder{tracer: provider.Tracer("riido-test")}
	ctx, span := recorder.StartTraceSpan(context.Background(), riidoaiserver.TraceSpanStart{
		Name: "riido.test.span",
		Kind: riidoaiserver.TraceSpanKindServer,
		Attributes: []riidoaiserver.TraceAttribute{
			riidoaiserver.StringTraceAttribute("riido.route", "/v1/test"),
		},
	})
	if ctx == nil {
		t.Fatal("StartTraceSpan returned nil context")
	}
	span.SetAttributes(riidoaiserver.BoolTraceAttribute("riido.allowed", true))
	span.RecordError(errors.New("boom"))
	span.End()

	spans := exporter.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("recorded spans = %d, want 1", len(spans))
	}
	got := spans[0]
	if got.Name() != "riido.test.span" || got.SpanKind() != trace.SpanKindServer {
		t.Fatalf("span = %q/%v", got.Name(), got.SpanKind())
	}
	if got.Status().Code != codes.Error || got.Status().Description != "boom" {
		t.Fatalf("span status = %v/%q", got.Status().Code, got.Status().Description)
	}
	attrs := map[string]attribute.Value{}
	for _, attr := range got.Attributes() {
		attrs[string(attr.Key)] = attr.Value
	}
	if attrs["riido.route"].AsString() != "/v1/test" || !attrs["riido.allowed"].AsBool() {
		t.Fatalf("span attrs = %#v", attrs)
	}
}
