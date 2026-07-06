package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestStartTraceSpanWithoutRecorderReturnsNoopSpan(t *testing.T) {
	ctx, span := StartTraceSpan(context.Background(), nil, TraceSpanStart{Name: "noop"})
	if ctx == nil || span == nil {
		t.Fatalf("noop tracing returned ctx=%v span=%v", ctx, span)
	}
	span.SetAttributes(StringTraceAttribute("key", "value"))
	span.RecordError(errors.New("ignored"))
	span.End()
	FinishTraceSpan(nil, errors.New("ignored"))
}

func TestFinishTraceSpanSuppressesCallerCancellation(t *testing.T) {
	canceled := &recordingTraceSpan{Attributes: map[string]string{}}
	FinishTraceSpan(canceled, context.Canceled)
	if len(canceled.Errors) != 0 || !canceled.Ended {
		t.Fatalf("canceled span = %+v", canceled)
	}
	failed := &recordingTraceSpan{Attributes: map[string]string{}}
	FinishTraceSpan(failed, errors.New("boom"))
	if len(failed.Errors) != 1 || failed.Errors[0] != "boom" || !failed.Ended {
		t.Fatalf("failed span = %+v", failed)
	}
}
