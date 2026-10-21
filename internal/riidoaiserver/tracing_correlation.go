package riidoaiserver

import "context"

// traceCorrelationReader exposes only validated OTel identifiers to structured logs.
type traceCorrelationReader interface {
	TraceCorrelation(context.Context) (string, string)
}

func traceCorrelationFromContext(ctx context.Context) (string, string) {
	reader, ok := TraceRecorderFromContext(ctx).(traceCorrelationReader)
	if !ok {
		return "", ""
	}
	traceID, spanID := reader.TraceCorrelation(ctx)
	if !validTraceHex(traceID, 32) || !validTraceHex(spanID, 16) {
		return "", ""
	}
	return traceID, spanID
}

func validTraceHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	nonzero := false
	for _, c := range value {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
		nonzero = nonzero || c != '0'
	}
	return nonzero
}
