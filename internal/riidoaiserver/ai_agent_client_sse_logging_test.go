package riidoaiserver

import (
	"bytes"
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
	logAIAgentClientStreamOpen(request, 7)
	logAIAgentClientStreamClose(request, "event_write_error", errors.New("broken pipe"))
	logAIAgentClientSubscriberDeliverySummary(aiAgentClientSubscriber{droppedEvents: 3, terminalCompensations: 1})
	got := output.String()
	for _, want := range []string{"event=stream_open", "event=stream_closed", "event=fanout_overflow_summary", "terminal_compensations=1", "broken pipe"} {
		if !strings.Contains(got, want) {
			t.Fatalf("SSE lifecycle log missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, "secret") {
		t.Fatalf("SSE lifecycle log exposed query: %s", got)
	}
}
