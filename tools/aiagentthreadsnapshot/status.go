package main

func isTerminal(state string) bool {
	switch state {
	case "completed", "failed", "stopped", "cancelled", "timeout":
		return true
	default:
		return false
	}
}

func decide(rep report) decisionSummary {
	if hasEndpointError(rep.Endpoints) {
		return decisionSummary{Status: "partial", Reason: "one or more API endpoints failed"}
	}
	if hasTerminalLiveConflict(rep) {
		return decisionSummary{Status: "captured_terminal_live_conflict", Reason: "terminal thread has active stream or matching live SSE event"}
	}
	if rep.V3.RunningCount > 0 || rep.V2.RunningCount > 0 || len(rep.SSEEvents) > 0 {
		return decisionSummary{Status: "captured", Reason: "snapshot includes active or live progress signal"}
	}
	return decisionSummary{Status: "baseline", Reason: "no active progress signal observed during capture window"}
}

func hasEndpointError(endpoints []endpointObservation) bool {
	for _, endpoint := range endpoints {
		if endpoint.Error != "" || endpoint.StatusCode == 0 {
			return true
		}
	}
	return false
}

func hasTerminalLiveConflict(rep report) bool {
	for _, thread := range append(rep.V3.HighlightedThreads, rep.V2.HighlightedThreads...) {
		if !isTerminal(thread.AssignmentState) {
			continue
		}
		if thread.ActiveStream || matchesLiveEvent(thread, rep.SSEEvents) {
			return true
		}
	}
	return false
}

func matchesLiveEvent(thread threadSurface, events []sseEventSummary) bool {
	for _, event := range events {
		if event.LineCount == 0 && event.WorkStatus == "" && event.AssignmentState == "" {
			continue
		}
		if sameNonEmpty(thread.AssignmentID, event.AssignmentID) ||
			sameNonEmpty(thread.RunID, event.RunID) ||
			sameNonEmpty(thread.ThreadID, event.ThreadID) {
			return true
		}
	}
	return false
}

func sameNonEmpty(a, b string) bool {
	return a != "" && a == b
}
