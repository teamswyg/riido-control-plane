package riidoaiserver

import "context"

// A connection span may remain open for hours. End each observation separately
// so its log can be joined to an exported span before that connection closes.
func startSSEObservation(ctx context.Context, event string, attributes ...TraceAttribute) (context.Context, TraceSpan) {
	return StartTraceSpan(ctx, nil, TraceSpanStart{
		Name: "SSE " + event, Kind: TraceSpanKindInternal,
		Attributes: append([]TraceAttribute{StringTraceAttribute("riido.trace.surface", "client_sse"), StringTraceAttribute("riido.sse.event", event)}, attributes...),
	})
}
