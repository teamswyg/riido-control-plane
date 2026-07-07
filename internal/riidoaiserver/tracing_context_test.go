package riidoaiserver

import (
	"context"
	"testing"
)

func TestTraceRecorderContextEdges(t *testing.T) {
	if got := TraceRecorderFromContext(nilTraceContext()); got != nil {
		t.Fatalf("nil context recorder = %T, want nil", got)
	}
	base := WithTraceRecorder(nilTraceContext(), nil)
	if base == nil {
		t.Fatal("nil context with nil recorder returned nil")
	}
	recorder := &recordingTraceRecorder{}
	ctx := WithTraceRecorder(nilTraceContext(), recorder)
	if got := TraceRecorderFromContext(ctx); got != recorder {
		t.Fatalf("context recorder = %T, want %T", got, recorder)
	}
}

func TestStartTraceSpanUsesExplicitRecorderWithNilContext(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	ctx, span := StartTraceSpan(nilTraceContext(), recorder, TraceSpanStart{
		Name: "explicit",
		Kind: TraceSpanKindClient,
		Attributes: []TraceAttribute{
			Int64TraceAttribute("attempt", 2),
			BoolTraceAttribute("cache_hit", true),
		},
	})
	if ctx == nil || span == nil {
		t.Fatalf("span ctx=%v span=%v", ctx, span)
	}
	span.End()
	spans := recorder.snapshot()
	if len(spans) != 1 || spans[0].Name != "explicit" {
		t.Fatalf("spans = %+v", spans)
	}
	if spans[0].Attributes["attempt"] != "2" || spans[0].Attributes["cache_hit"] != "true" {
		t.Fatalf("attributes = %+v", spans[0].Attributes)
	}
}

func TestStartTraceSpanUsesRecorderFromContext(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	ctx := WithTraceRecorder(context.Background(), recorder)
	next, span := StartTraceSpan(ctx, nil, TraceSpanStart{Name: "context"})
	if next == nil || span == nil {
		t.Fatalf("span ctx=%v span=%v", next, span)
	}
	span.SetAttributes(StringTraceAttribute("source", "ctx"))
	span.End()
	spans := recorder.snapshot()
	if len(spans) != 1 || spans[0].Attributes["source"] != "ctx" || !spans[0].Ended {
		t.Fatalf("spans = %+v", spans)
	}
}

func TestTaskContextOperationNameStringFallback(t *testing.T) {
	if got := TaskContextOperationResolve.String(); got != "task_context_resolve" {
		t.Fatalf("operation = %q", got)
	}
	if got := (TaskContextOperationName(" \t")).String(); got != unknownTaskContextOperation {
		t.Fatalf("blank operation = %q, want %q", got, unknownTaskContextOperation)
	}
}

func nilTraceContext() context.Context {
	return nil
}
