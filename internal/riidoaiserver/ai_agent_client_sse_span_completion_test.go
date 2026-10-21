package riidoaiserver

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type sseEventSpanContextKey struct{}

type sseEventSpanRecorder struct{ recordingTraceRecorder }

func (r *sseEventSpanRecorder) StartTraceSpan(ctx context.Context, start TraceSpanStart) (context.Context, TraceSpan) {
	ctx, span := r.recordingTraceRecorder.StartTraceSpan(ctx, start)
	return context.WithValue(ctx, sseEventSpanContextKey{}, true), span
}

func (*sseEventSpanRecorder) TraceCorrelation(ctx context.Context) (string, string) {
	if child, _ := ctx.Value(sseEventSpanContextKey{}).(bool); child {
		return "12345678901234567890123456789012", "2222222222222222"
	}
	return "12345678901234567890123456789012", "1111111111111111"
}

func TestSSELifecycleLogReferencesCompletedEventSpanWhileConnectionRemainsOpen(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	recorder := &sseEventSpanRecorder{}
	ctx := WithTraceRecorder(context.Background(), recorder)
	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/private-workspace/ai-agent/events?token=private-token", nil).WithContext(ctx)
	logAIAgentClientStreamOpen(req, 0)
	logAIAgentClientTerminalDelivery("live_written", ClientStreamEvent{Payload: AgentThreadProgressEvent{AssignmentState: AgentAssignmentStateCompleted}}, "private-workspace", ctx)
	logAIAgentClientStreamClose(req, "event_write_error", errors.New("private-response"))
	spans := recorder.snapshot()
	if len(spans) != 3 {
		t.Fatalf("completed stream observation spans = %d, want 3", len(spans))
	}
	for _, span := range spans {
		if !span.Ended {
			t.Fatalf("observation %q waits for the HTTP SSE span to end", span.Name)
		}
	}
	if strings.Count(output.String(), `span_id="2222222222222222"`) != 3 || strings.Contains(output.String(), `span_id="1111111111111111"`) {
		t.Fatal("SSE logs point at the unfinished connection span instead of their completed event span")
	}
	for _, private := range []string{"private-workspace", "private-token", "private-response"} {
		if strings.Contains(output.String(), private) {
			t.Fatal("stream observation leaked private input")
		}
	}
	if len(spans[2].Errors) != 1 || spans[2].Errors[0] != "SSE write failed" {
		t.Fatal("write failure did not retain a fixed error signal")
	}
}
