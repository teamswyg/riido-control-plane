package main

func summarizeSubscription(payload subscriptionPayload, conversationID string) subscriptionSummary {
	summary := subscriptionSummary{
		StreamHref: payload.Stream.Href, EventType: payload.Stream.EventType,
		ActiveThreadFilterCount: len(payload.ActiveThreadFilters),
	}
	if conversationID == "" {
		return summary
	}
	for _, filter := range payload.ActiveThreadFilters {
		if filter.ThreadID == conversationID || filter.RunID == conversationID {
			summary.HighlightedFilterMatched = true
		}
	}
	return summary
}
