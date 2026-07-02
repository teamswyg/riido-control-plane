package main

func summarizeSubscription(
	payload subscriptionPayload,
	threads []threadSurface,
	conversationID string,
) subscriptionSummary {
	summary := subscriptionSummary{
		StreamHref: payload.Stream.Href, EventType: payload.Stream.EventType,
		ActiveThreadFilterCount: len(payload.ActiveThreadFilters),
	}
	for _, filter := range payload.ActiveThreadFilters {
		if filter.ThreadID == conversationID || filter.RunID == conversationID {
			summary.HighlightedFilterMatched = true
		}
		for _, thread := range threads {
			if !filterMatchesThread(filter, thread) {
				continue
			}
			summary.HighlightedFilterMatched = true
			if isTerminal(thread.AssignmentState) {
				summary.TerminalFilterMatched = true
			}
		}
	}
	return summary
}

func filterMatchesThread(filter threadFilter, thread threadSurface) bool {
	return sameNonEmpty(filter.ThreadID, thread.ThreadID) ||
		sameNonEmpty(filter.RunID, thread.RunID)
}
