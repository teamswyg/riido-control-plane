package main

import (
	"context"
	"net/http"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type otelTraceRecorder struct {
	tracer trace.Tracer
}

type otelTraceSpan struct {
	span trace.Span
}

func (r otelTraceRecorder) ExtractTraceContext(ctx context.Context, header http.Header) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

func (r otelTraceRecorder) InjectTraceContext(ctx context.Context, header http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
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
