package riidoaiserver

import (
	"log"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

// Logged after both durable event acceptance and client projection. Assignment
// age uses server timestamps only. This is not a client-render latency metric.
func logAgentResponseDelivery(req AgentEventRequest, response AgentEventResponse, receivedAt, committedAt, projectedAt time.Time) {
	phase := ""
	switch {
	case req.Metadata[metadatakeys.ProgressMessageKey.String()] == "assistant.partial":
		phase = "partial_projected"
	case isTerminal(req.State):
		phase = "terminal_projected"
	}
	if phase == "" {
		return
	}
	age := int64(-1)
	if response.Assignment != nil && !response.Assignment.CreatedAt.IsZero() && !projectedAt.Before(response.Assignment.CreatedAt) {
		age = projectedAt.Sub(response.Assignment.CreatedAt).Milliseconds()
	}
	log.Printf("event=agent_response_delivery assignment_id=%q phase=%q event_seq=%d event_processing_ms=%d assignment_commit_ms=%d client_projection_ms=%d assignment_age_ms=%d", req.AssignmentID, phase, response.Event.Seq, projectedAt.Sub(receivedAt).Milliseconds(), committedAt.Sub(receivedAt).Milliseconds(), projectedAt.Sub(committedAt).Milliseconds(), age)
}
