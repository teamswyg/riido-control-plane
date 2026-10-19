package riidoaiserver

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"
)

func TestResponseTimingSplitRemainsAdditiveAfterRounding(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	start := time.Unix(100, 0)
	logAgentResponseDelivery(AgentEventRequest{AssignmentID: "rounding", State: AssignmentCompleted}, AgentEventResponse{}, start, start.Add(1800*time.Microsecond), start.Add(2600*time.Microsecond))
	for _, want := range []string{"event_processing_ms=2", "assignment_commit_ms=1", "client_projection_ms=1"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %s: %s", want, output.String())
		}
	}
}
