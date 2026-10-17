package riidoaiserver

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestResponseTimingLogsOnlyAcceptedDeliveryMetadata(t *testing.T) {
	var out bytes.Buffer
	old := log.Writer()
	log.SetOutput(&out)
	defer log.SetOutput(old)
	received := time.Unix(100, 0)
	req := AgentEventRequest{AssignmentID: "assignment", Message: "PRIVATE ANSWER", Metadata: map[string]string{metadatakeys.ProgressMessageKey.String(): "assistant.partial", "token": "PRIVATE TOKEN"}}
	response := AgentEventResponse{Assignment: &Assignment{CreatedAt: received.Add(-2 * time.Second)}, Event: TaskEvent{Seq: 7}}
	logAgentResponseDelivery(req, response, received, received.Add(2*time.Millisecond), received.Add(5*time.Millisecond))
	got := out.String()
	for _, want := range []string{"partial_projected", "event_processing_ms=5", "assignment_commit_ms=2", "client_projection_ms=3", "assignment_age_ms=2005", "event_seq=7"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %s: %s", want, got)
		}
	}
	if strings.Contains(got, "PRIVATE") {
		t.Fatal("content leaked")
	}
	out.Reset()
	req.Metadata = nil
	logAgentResponseDelivery(req, response, received, received, received)
	if out.Len() != 0 {
		t.Fatal("unrelated event logged")
	}
	req.State = AssignmentFailed
	response.Assignment = nil
	logAgentResponseDelivery(req, response, received, received, received)
	if !strings.Contains(out.String(), "assignment_age_ms=-1") {
		t.Fatal("missing timestamp treated as zero")
	}
}
