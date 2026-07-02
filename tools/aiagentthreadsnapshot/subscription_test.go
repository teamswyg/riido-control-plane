package main

import "testing"

func TestSummarizeSubscriptionMarksTerminalFilter(t *testing.T) {
	payload := subscriptionPayload{ActiveThreadFilters: []threadFilter{
		{ThreadID: "thread-1", RunID: "run-1"},
	}}
	threads := []threadSurface{{
		ThreadID: "thread-1", RunID: "run-1", AssignmentState: "completed",
	}}
	got := summarizeSubscription(payload, threads, "")
	if !got.HighlightedFilterMatched || !got.TerminalFilterMatched {
		t.Fatalf("subscription summary missed terminal filter: %+v", got)
	}
}
