package riidoaiserver

import (
	"context"
	"testing"
)

func TestTraceCorrelationRejectsInvalidOrMissingIdentifiers(t *testing.T) {
	for _, value := range []string{"", "0000000000000000", "0123456789ABCDEF", "Bearer credential", "/Users/private/path"} {
		if validTraceHex(value, 16) {
			t.Fatalf("untrusted trace ID accepted: %q", value)
		}
	}
	if traceID, spanID := traceCorrelationFromContext(context.Background()); traceID != "" || spanID != "" {
		t.Fatal("invented correlation")
	}
}
