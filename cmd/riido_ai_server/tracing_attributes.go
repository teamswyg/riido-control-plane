package main

import (
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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
		switch attr.Kind {
		case riidoaiserver.TraceAttributeKindInt64:
			out = append(out, attribute.Int64(key, attr.Int64Value))
		case riidoaiserver.TraceAttributeKindBool:
			out = append(out, attribute.Bool(key, attr.BoolValue))
		default:
			out = append(out, attribute.String(key, attr.Value))
		}
	}
	return out
}
