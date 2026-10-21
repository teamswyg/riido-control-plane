package riidoaiserver

import (
	"context"
	"errors"
	"time"
)

type taskContextReadError struct{ outcome string }

func (e taskContextReadError) Error() string { return "task context read " + e.outcome }

func taskContextReadOutcome(ctx context.Context, err error) string {
	if err == nil {
		return "available"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "budget_exhausted"
	}
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return "cancelled"
	}
	var failure taskContextReadError
	if errors.As(err, &failure) {
		return failure.outcome
	}
	return "unavailable"
}

func observeTaskContextRead(ctx context.Context, operation string) (context.Context, func(string, ...TraceAttribute)) {
	started := time.Now()
	ctx, span := StartTraceSpan(ctx, nil, TraceSpanStart{Name: "task_context." + operation, Kind: TraceSpanKindClient})
	return ctx, func(outcome string, attributes ...TraceAttribute) {
		span.SetAttributes(attributes...)
		span.SetAttributes(StringTraceAttribute("riido.task_context.operation", operation), StringTraceAttribute("riido.task_context.outcome", outcome), Int64TraceAttribute("riido.task_context.duration_ms", time.Since(started).Milliseconds()))
		var failure error
		if outcome != "available" && outcome != "empty" && outcome != "partial" {
			failure = errors.New("task context read unavailable")
		}
		FinishTraceSpan(span, failure)
	}
}
