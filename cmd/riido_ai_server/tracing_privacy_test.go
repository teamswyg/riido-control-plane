package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOtelTraceErrorDoesNotExportPrivateMessages(t *testing.T) {
	const marker = "FAKE_PRIVATE_RESPONSE_MARKER"
	cases := []struct {
		name string
		err  error
		kind string
	}{
		{"upstream response", errors.New(`dynamodb api error: status=400 body="` + marker + `"`), "operation_failed"},
		{"wrapped deadline", fmt.Errorf("%s: %w", marker, context.DeadlineExceeded), "deadline_exceeded"},
		{"network timeout", &url.Error{Op: "POST", URL: "https://unit.invalid/" + marker, Err: &net.DNSError{Err: marker, IsTimeout: true}}, "timeout"},
		{"forbidden", fmt.Errorf("%s: %w", marker, riidoaiserver.ErrAuthorizationForbidden), "forbidden"},
		{"unauthenticated", fmt.Errorf("%s: %w", marker, riidoaiserver.ErrAuthorizationUnauthenticated), "unauthenticated"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exporter := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
			_, span := (otelTraceRecorder{tracer: provider.Tracer("privacy-test")}).StartTraceSpan(context.Background(), riidoaiserver.TraceSpanStart{Name: "operation"})
			span.RecordError(tc.err)
			span.End()
			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatalf("spans = %d", len(spans))
			}
			body, err := json.Marshal(spans)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(body), marker) {
				t.Fatal("private error message crossed the OTel boundary")
			}
			got := spans.Snapshots()[0]
			if got.Status().Code != codes.Error || got.Status().Description != "operation failed" {
				t.Fatal("safe failure status was not retained")
			}
			var kind string
			for _, attr := range got.Attributes() {
				if string(attr.Key) == "error.type" {
					kind = attr.Value.AsString()
				}
			}
			if kind != tc.kind {
				t.Fatalf("bounded error type = %q, want %q", kind, tc.kind)
			}
			if len(got.Events()) != 1 {
				t.Fatal("failure exception signal was lost")
			}
		})
	}
}

type messageForbiddenTraceError struct{}

func (messageForbiddenTraceError) Error() string { panic("raw Error() must not be invoked") }

func TestOtelTraceErrorDoesNotEvaluateArbitraryMessage(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	_, span := (otelTraceRecorder{tracer: provider.Tracer("privacy-test")}).StartTraceSpan(context.Background(), riidoaiserver.TraceSpanStart{Name: "operation"})
	span.RecordError(messageForbiddenTraceError{})
	span.End()
	if spans := exporter.GetSpans(); len(spans) != 1 || spans.Snapshots()[0].Status().Code != codes.Error {
		t.Fatal("safe error signal was not recorded")
	}
}
