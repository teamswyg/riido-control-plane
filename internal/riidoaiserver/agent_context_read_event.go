package riidoaiserver

import (
	"context"
	"time"
)

// AgentClientEventContextRead streams "the agent read these sources" to the
// client while a run is in progress, from the same records the server keeps.
const AgentClientEventContextRead = "agent_context_read"

// agentContextReadEventMaxSources keeps one event small inside the shared
// client event replay buffer; the full list stays in the read records.
const agentContextReadEventMaxSources = 32

type AgentContextReadSource struct {
	Kind      string `json:"kind"`
	ID        string `json:"id,omitempty"`
	Part      string `json:"part"`
	Version   string `json:"version,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Status    string `json:"status"`
}

// AgentContextReadEvent carries one served read (all sources of one prompt
// capture or one CLI call). Requester identity and CLI request ids stay in
// the server-side records only.
type AgentContextReadEvent struct {
	EventType      string                   `json:"event_type"`
	SchemaVersion  string                   `json:"schema_version"`
	AgentID        string                   `json:"agent_id"`
	TaskID         string                   `json:"task_id"`
	AssignmentID   string                   `json:"assignment_id"`
	ThreadID       string                   `json:"thread_id"`
	RunID          string                   `json:"run_id"`
	Via            string                   `json:"via"`
	TriggerKind    string                   `json:"trigger_kind,omitempty"`
	ServedAt       time.Time                `json:"served_at"`
	Sources        []AgentContextReadSource `json:"sources"`
	OmittedSources int                      `json:"omitted_sources,omitempty"`
	// Cut-off bookkeeping for per-subscriber counts; never serialized.
	omittedAgentIDs []string
	omittedOther    int
}

// AIAgentContextReadRecorder publishes read records to the client stream.
type AIAgentContextReadRecorder interface {
	RecordAIAgentContextReads(ctx context.Context, reads []AssignmentContextRead) error
}

// newAgentContextReadEvent builds the event without thread identity; reads
// must share one assignment, agent and serve call.
func newAgentContextReadEvent(reads []AssignmentContextRead) AgentContextReadEvent {
	first := reads[0]
	event := AgentContextReadEvent{
		EventType: AgentClientEventContextRead, SchemaVersion: SchemaVersion,
		AgentID: first.AgentID, TaskID: first.TaskID, AssignmentID: first.AssignmentID,
		Via: first.Via, TriggerKind: first.TriggerKind, ServedAt: first.ServedAt.UTC(),
		Sources: make([]AgentContextReadSource, 0, min(len(reads), agentContextReadEventMaxSources)),
	}
	for _, read := range reads {
		if len(event.Sources) == agentContextReadEventMaxSources {
			event.noteOmitted(read)
			continue
		}
		event.Sources = append(event.Sources, AgentContextReadSource{
			Kind: read.SourceKind, ID: read.SourceID, Part: read.SourcePart,
			Version: read.SourceVersion, UpdatedAt: read.SourceUpdatedAt, Status: read.Status,
		})
	}
	return event
}
