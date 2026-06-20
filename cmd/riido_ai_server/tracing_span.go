package main

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel/codes"
)

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
