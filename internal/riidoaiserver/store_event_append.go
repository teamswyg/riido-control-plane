package riidoaiserver

import (
	"context"
	"time"
)

func (s *Store) appendEvent(state *storeState, taskID, assignmentID, agentID, eventType string, assignmentState AssignmentState, message string, metadata map[string]string, at time.Time) TaskEvent {
	startedAt := time.Now()
	defer func() {
		recordEventAppendLatency(state, time.Since(startedAt))
	}()
	state.nextEventSeq++
	event := TaskEvent{
		Seq:          state.nextEventSeq,
		TaskID:       taskID,
		AssignmentID: assignmentID,
		AgentID:      agentID,
		Type:         eventType,
		State:        assignmentState,
		Message:      message,
		Metadata:     cloneMetadata(metadata),
		At:           at,
	}
	state.events[taskID] = append(state.events[taskID], event)
	if s.outbox != nil {
		s.appendTaskEventToOutbox(state, event)
	}
	for _, ch := range state.subscribers[taskID] {
		select {
		case ch <- event:
		default:
		}
	}
	return event
}

func (s *Store) appendTaskEventToOutbox(state *storeState, event TaskEvent) {
	err := s.outbox.AppendTaskEvent(context.Background(), event)
	if err != nil {
		state.outboxErrorsTotal++
	}
}

func recordEventAppendLatency(state *storeState, duration time.Duration) {
	milliseconds := durationMilliseconds(duration)
	state.eventAppendLatency.samplesTotal++
	state.eventAppendLatency.totalMilliseconds += milliseconds
	state.eventAppendLatency.lastMilliseconds = milliseconds
	if milliseconds > state.eventAppendLatency.maxMilliseconds {
		state.eventAppendLatency.maxMilliseconds = milliseconds
	}
}
