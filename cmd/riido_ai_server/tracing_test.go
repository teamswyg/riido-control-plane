package main

import (
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
	riidoaiserver "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestOtelAttributesPreserveTypedValues(t *testing.T) {
	attrs := otelAttributes([]riidoaiserver.TraceAttribute{
		riidoaiserver.StringTraceAttribute(metadatakeys.HTTPRoute.String(), "/v1/daemon/runtime-snapshot"),
		riidoaiserver.Int64TraceAttribute(metadatakeys.HTTPStatusCode.String(), 202),
		riidoaiserver.BoolTraceAttribute("riido.test.sampled", true),
	})

	got := map[string]attribute.Value{}
	for _, attr := range attrs {
		got[string(attr.Key)] = attr.Value
	}
	if got[metadatakeys.HTTPRoute.String()].Type() != attribute.STRING {
		t.Fatalf("http route type = %s", got[metadatakeys.HTTPRoute.String()].Type())
	}
	if got[metadatakeys.HTTPStatusCode.String()].Type() != attribute.INT64 ||
		got[metadatakeys.HTTPStatusCode.String()].AsInt64() != 202 {
		t.Fatalf("http status value = %s/%d", got[metadatakeys.HTTPStatusCode.String()].Type(), got[metadatakeys.HTTPStatusCode.String()].AsInt64())
	}
	if got["riido.test.sampled"].Type() != attribute.BOOL || !got["riido.test.sampled"].AsBool() {
		t.Fatalf("bool value = %s/%v", got["riido.test.sampled"].Type(), got["riido.test.sampled"].AsBool())
	}
}

func TestOtelSpanKindMapsServerClientAndDefault(t *testing.T) {
	tests := []struct {
		name string
		kind riidoaiserver.TraceSpanKind
		want trace.SpanKind
	}{
		{name: "server", kind: riidoaiserver.TraceSpanKindServer, want: trace.SpanKindServer},
		{name: "client", kind: riidoaiserver.TraceSpanKindClient, want: trace.SpanKindClient},
		{name: "internal", kind: riidoaiserver.TraceSpanKindInternal, want: trace.SpanKindInternal},
		{name: "unknown", kind: riidoaiserver.TraceSpanKind(99), want: trace.SpanKindInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := otelSpanKind(tt.kind); got != tt.want {
				t.Fatalf("otelSpanKind(%v) = %v, want %v", tt.kind, got, tt.want)
			}
		})
	}
}
