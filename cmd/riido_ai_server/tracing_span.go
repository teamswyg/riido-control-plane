package main

import (
	"context"
	"errors"
	"net"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (s otelTraceSpan) RecordError(err error) {
	if s.span == nil || err == nil || errors.Is(err, context.Canceled) {
		return
	}
	// Errors can carry upstream response bodies and private URLs. Keep the original
	// error for application control flow; export only a fixed message and category.
	s.span.SetAttributes(attribute.String("error.type", safeTraceErrorType(err)))
	s.span.RecordError(errors.New("operation failed"))
	s.span.SetStatus(codes.Error, "operation failed")
}

func safeTraceErrorType(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, riidoaiserver.ErrAuthorizationForbidden):
		return "forbidden"
	case errors.Is(err, riidoaiserver.ErrAuthorizationUnauthenticated):
		return "unauthenticated"
	case errors.Is(err, riidoaiserver.ErrAuthorizationInvalidCredential):
		return "invalid_credential"
	default:
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			return "timeout"
		}
		return "operation_failed"
	}
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
