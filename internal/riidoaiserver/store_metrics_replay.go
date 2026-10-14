package riidoaiserver

func rebuildStateMetricsFromHistory(state *storeState) {
	if state == nil {
		return
	}
	replayed := replayedMetricsFromEvents(state.events)
	if replayed.pollStartTotal > state.pollActionsTotal[PollStart] {
		state.pollActionsTotal[PollStart] = replayed.pollStartTotal
	}
	if replayed.agentEventsTotal > state.agentEventsTotal {
		state.agentEventsTotal = replayed.agentEventsTotal
	}
	if replayed.taskEventsTotal > state.eventAppendLatency.samplesTotal {
		state.eventAppendLatency.samplesTotal = replayed.taskEventsTotal
	}
	repairEventAppendLatencyFromHistory(state)
	pollActionsTotal := int64(0)
	for _, count := range state.pollActionsTotal {
		pollActionsTotal += count
	}
	if pollActionsTotal > state.pollRequestsTotal {
		state.pollRequestsTotal = pollActionsTotal
	}
}

func repairEventAppendLatencyFromHistory(state *storeState) {
	if state == nil || state.eventAppendLatency.samplesTotal <= 0 {
		return
	}
	// Historical snapshots and operation replays know that events were appended
	// but do not know the original wall-clock duration. Keep the metric useful
	// by applying the same 1ms minimum used by live sub-millisecond samples.
	minimumTotal := state.eventAppendLatency.samplesTotal
	if state.eventAppendLatency.totalMilliseconds < minimumTotal {
		state.eventAppendLatency.totalMilliseconds = minimumTotal
	}
	if state.eventAppendLatency.maxMilliseconds <= 0 {
		state.eventAppendLatency.maxMilliseconds = 1
	}
	if state.eventAppendLatency.lastMilliseconds <= 0 {
		state.eventAppendLatency.lastMilliseconds = 1
	}
}

type replayedMetrics struct {
	taskEventsTotal  int64
	pollStartTotal   int64
	agentEventsTotal int64
}

func replayedMetricsFromEvents(events map[string][]TaskEvent) replayedMetrics {
	metrics := replayedMetrics{}
	for _, taskEvents := range events {
		for _, event := range taskEvents {
			if event.Seq <= 0 || event.TaskID == "" {
				continue
			}
			metrics.taskEventsTotal++
			if event.Type == EventAssignmentLeased {
				metrics.pollStartTotal++
			}
			if eventLooksAgentAuthored(event) {
				metrics.agentEventsTotal++
			}
		}
	}
	return metrics
}

func eventLooksAgentAuthored(event TaskEvent) bool {
	switch event.Type {
	case EventAssignmentReady,
		EventAssignmentRunning,
		EventAssignmentCompleted,
		EventAssignmentFailed,
		EventAssignmentStateUpdated,
		EventRiidoLog,
		EventProviderLog,
		EventProviderWarning,
		EventProviderError:
		return true
	default:
		return false
	}
}
