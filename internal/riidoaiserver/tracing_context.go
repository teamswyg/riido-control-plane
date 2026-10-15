package riidoaiserver

import (
	"context"
	"net/http"
)

type traceRecorderContextKey struct{}

func WithTraceRecorder(ctx context.Context, recorder TraceRecorder) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if recorder == nil {
		return ctx
	}
	return context.WithValue(ctx, traceRecorderContextKey{}, recorder)
}

func TraceRecorderFromContext(ctx context.Context) TraceRecorder {
	if ctx == nil {
		return nil
	}
	recorder, _ := ctx.Value(traceRecorderContextKey{}).(TraceRecorder)
	return recorder
}

func ExtractTraceContext(ctx context.Context, recorder TraceRecorder, header http.Header) context.Context {
	if propagator, ok := recorder.(traceContextPropagator); ok {
		return propagator.ExtractTraceContext(ctx, header)
	}
	return ctx
}

func InjectTraceContext(ctx context.Context, header http.Header) {
	if propagator, ok := TraceRecorderFromContext(ctx).(traceContextPropagator); ok {
		propagator.InjectTraceContext(ctx, header)
	}
}
