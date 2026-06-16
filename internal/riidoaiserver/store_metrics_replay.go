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
	pollActionsTotal := int64(0)
	for _, count := range state.pollActionsTotal {
		pollActionsTotal += count
	}
	if pollActionsTotal > state.pollRequestsTotal {
		state.pollRequestsTotal = pollActionsTotal
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
