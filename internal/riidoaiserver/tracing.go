package riidoaiserver

import (
	"context"
	"errors"
	"strconv"
)

type TraceSpanKind int

const (
	TraceSpanKindInternal TraceSpanKind = iota
	TraceSpanKindServer
	TraceSpanKindClient
)

type TraceAttribute struct {
	Key        string
	Value      string
	Int64Value int64
	BoolValue  bool
	Kind       TraceAttributeKind
}

type TraceAttributeKind int

const (
	TraceAttributeKindString TraceAttributeKind = iota
	TraceAttributeKindInt64
	TraceAttributeKindBool
)

func StringTraceAttribute(key, value string) TraceAttribute {
	return TraceAttribute{Key: key, Value: value}
}

func Int64TraceAttribute(key string, value int64) TraceAttribute {
	return TraceAttribute{Key: key, Int64Value: value, Kind: TraceAttributeKindInt64}
}

func BoolTraceAttribute(key string, value bool) TraceAttribute {
	return TraceAttribute{Key: key, BoolValue: value, Kind: TraceAttributeKindBool}
}

func (attr TraceAttribute) StringValue() string {
	switch attr.Kind {
	case TraceAttributeKindInt64:
		return strconv.FormatInt(attr.Int64Value, 10)
	case TraceAttributeKindBool:
		return strconv.FormatBool(attr.BoolValue)
	default:
		return attr.Value
	}
}

type TraceSpanStart struct {
	Name       string
	Kind       TraceSpanKind
	Attributes []TraceAttribute
}

type TraceSpan interface {
	SetAttributes(attributes ...TraceAttribute)
	RecordError(err error)
	End()
}

type TraceRecorder interface {
	StartTraceSpan(ctx context.Context, start TraceSpanStart) (context.Context, TraceSpan)
}

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

func StartTraceSpan(ctx context.Context, recorder TraceRecorder, start TraceSpanStart) (context.Context, TraceSpan) {
	if ctx == nil {
		ctx = context.Background()
	}
	if recorder == nil {
		recorder = TraceRecorderFromContext(ctx)
	}
	if recorder == nil {
		return ctx, noopTraceSpan{}
	}
	ctx = WithTraceRecorder(ctx, recorder)
	return recorder.StartTraceSpan(ctx, start)
}

func FinishTraceSpan(span TraceSpan, err error) {
	if span == nil {
		return
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		span.RecordError(err)
	}
	span.End()
}

type noopTraceSpan struct{}

func (noopTraceSpan) SetAttributes(attributes ...TraceAttribute) { _ = len(attributes) }

func (noopTraceSpan) RecordError(err error) { _ = err }

func (noopTraceSpan) End() { _ = TraceSpanKindInternal }
