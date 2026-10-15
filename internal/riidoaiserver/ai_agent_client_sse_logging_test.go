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

func TestAIAgentClientSSELifecycleLogsReasonWithoutQuery(t *testing.T) {
	var output bytes.Buffer
	previousWriter := log.Writer()
	previousFlags := log.Flags()
	log.SetOutput(&output)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
	}()

	request := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/ws/ai-agent/events?token=secret", nil)
	request.Header.Set("Last-Event-ID", "42")
	logAIAgentClientStreamOpen(request, 7)
	logAIAgentClientStreamClose(request, "event_write_error", errors.New("broken pipe"))
	logAIAgentClientSubscriberDeliverySummary(aiAgentClientSubscriber{droppedEvents: 3, terminalCompensations: 1})
	logAIAgentClientTerminalDelivery("live_written", ClientStreamEvent{Payload: AgentThreadProgressEvent{
		AssignmentID: "asn-terminal", ThreadID: "thread-terminal", RunID: "run-terminal",
		AssignmentState: AgentAssignmentStateCompleted,
	}}, "ws")
	got := output.String()
	for _, want := range []string{"event=stream_open", "replay_cursor_present=true", "replay_events=7", "event=stream_closed", "event=fanout_overflow_summary", "event=terminal_delivery", "assignment_ref=", "terminal_compensations=1"} {
		if !strings.Contains(got, want) {
			t.Fatalf("SSE lifecycle log missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, "secret") || strings.Contains(got, "broken pipe") || strings.Contains(got, "/workspaces/ws/") || strings.Contains(got, "asn-terminal") || strings.Contains(got, "thread-terminal") || strings.Contains(got, "run-terminal") {
		t.Fatalf("SSE lifecycle log exposed query: %s", got)
	}
}

type sseCorrelationRecorder struct{ recordingTraceRecorder }

func (s *sseCorrelationRecorder) TraceCorrelation(context.Context) (string, string) {
	return "12345678901234567890123456789012", "1234567890123456"
}

func TestAIAgentClientSSELifecycleRejectsRawPathReasonAndError(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	request := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/customer-secret/ai-agent/events?token=credential", nil)
	request = request.WithContext(WithTraceRecorder(request.Context(), &sseCorrelationRecorder{}))
	logAIAgentClientStreamOpen(request, 0)
	logAIAgentClientStreamClose(request, "/Users/private/customer", errors.New("Bearer credential customer-secret"))
	got := output.String()
	for _, forbidden := range []string{"customer-secret", "credential", "/Users/", "Bearer"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("SSE lifecycle log leaked sensitive input: %q", got)
		}
	}
	for _, want := range []string{`route="/v2/client/workspaces/{workspace_id}/ai-agent/events"`, `reason="unknown"`, `error_type="unknown"`, `trace_id="12345678901234567890123456789012"`, `span_id="1234567890123456"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("SSE bounded correlation missing %q: %q", want, got)
		}
	}
}
