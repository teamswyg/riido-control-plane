package riidoaiserver

import (
	"maps"
	"time"
)

func (s *Store) handleMetrics(state *storeState) MetricsSnapshot {
	rebuildStateMetricsFromHistory(state)
	assignmentsByState := map[AssignmentState]int{}
	for _, assignment := range state.assignments {
		assignmentsByState[assignment.State]++
	}
	pollActions := make(map[PollAction]int64, len(state.pollActionsTotal))
	maps.Copy(pollActions, state.pollActionsTotal)
	snapshot := MetricsSnapshot{
		SchemaVersion:                       MetricsSchemaVersion,
		GeneratedAt:                         s.now(),
		TasksTotal:                          len(state.tasks),
		AssignmentsTotal:                    len(state.assignments),
		AssignmentsByState:                  assignmentsByState,
		PollRequestsTotal:                   state.pollRequestsTotal,
		PollActionsTotal:                    pollActions,
		AgentEventsTotal:                    state.agentEventsTotal,
		TaskEventsTotal:                     countTaskEvents(state),
		SSESubscribers:                      countSubscribers(state),
		OutboxErrorsTotal:                   state.outboxErrorsTotal,
		EventAppendLatencySamplesTotal:      state.eventAppendLatency.samplesTotal,
		EventAppendLatencyTotalMilliseconds: state.eventAppendLatency.totalMilliseconds,
		EventAppendLatencyMaxMilliseconds:   state.eventAppendLatency.maxMilliseconds,
		EventAppendLatencyLastMilliseconds:  state.eventAppendLatency.lastMilliseconds,
	}
	return s.operationMetrics.ApplyToMetricsSnapshot(snapshot)
}

func (s *Store) observeStoreOperation(operation StoreOperationName, startedAt time.Time, err error) {
	if s == nil || s.operationMetrics == nil {
		return
	}
	s.operationMetrics.ObserveStoreOperation(StoreOperationObservation{
		Operation: operation,
		Duration:  time.Since(startedAt),
		Err:       err,
	})
}

func countTaskEvents(state *storeState) int64 {
	var total int64
	for _, events := range state.events {
		total += int64(len(events))
	}
	return total
}

func countSubscribers(state *storeState) int {
	total := 0
	for _, subs := range state.subscribers {
		total += len(subs)
	}
	return total
}
