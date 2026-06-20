package riidoaiserver

import (
	"sort"
	"time"
)

func eventsAfterSeq(state *storeState, seq int64) []TaskEvent {
	var events []TaskEvent
	for _, taskEvents := range state.events {
		for _, event := range taskEvents {
			if event.Seq > seq {
				events = append(events, event)
			}
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Seq < events[j].Seq })
	return events
}

func (s *Store) appendRecordedEvent(state *storeState, event TaskEvent) {
	startedAt := time.Now()
	defer func() {
		recordEventAppendLatency(state, time.Since(startedAt))
	}()
	if event.Seq <= 0 || event.TaskID == "" {
		return
	}
	for _, existing := range state.events[event.TaskID] {
		if existing.Seq == event.Seq {
			return
		}
	}
	recorded := event
	recorded.Metadata = cloneMetadata(event.Metadata)
	state.events[recorded.TaskID] = append(state.events[recorded.TaskID], recorded)
	sort.Slice(state.events[recorded.TaskID], func(i, j int) bool {
		return state.events[recorded.TaskID][i].Seq < state.events[recorded.TaskID][j].Seq
	})
	if recorded.Seq > state.nextEventSeq {
		state.nextEventSeq = recorded.Seq
	}
	if recorded.AssignmentID != "" && recorded.State != "" {
		assignment := state.assignments[recorded.AssignmentID]
		if assignment.ID != "" {
			assignment.State = recorded.State
			if !recorded.At.IsZero() {
				assignment.UpdatedAt = recorded.At
			}
			state.assignments[assignment.ID] = assignment
		}
	}
	if s.outbox != nil {
		s.appendTaskEventToOutbox(state, recorded)
	}
	for _, ch := range state.subscribers[recorded.TaskID] {
		select {
		case ch <- recorded:
		default:
		}
	}
}
