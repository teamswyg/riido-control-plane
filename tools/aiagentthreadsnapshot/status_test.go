package main

import "testing"

func TestDecideCapturesTerminalLiveConflict(t *testing.T) {
	rep := report{
		V3: threadSummary{HighlightedThreads: []threadSurface{{
			ThreadID: "thread-1", AssignmentID: "asn-1",
			AssignmentState: "completed",
		}}},
		SSEEvents: []sseEventSummary{{
			ThreadID: "thread-1", AssignmentID: "asn-1",
			WorkStatus: "running", AssignmentState: "running", LineCount: 1,
		}},
	}
	got := decide(rep)
	if got.Status != "captured_terminal_live_conflict" {
		t.Fatalf("unexpected decision: %+v", got)
	}
}

func TestDecideKeepsHistoricalTerminalProgressBaseline(t *testing.T) {
	rep := report{
		V3: threadSummary{HighlightedThreads: []threadSurface{{
			ThreadID: "thread-1", AssignmentID: "asn-1",
			AssignmentState: "completed", LineCount: 3,
		}}},
	}
	got := decide(rep)
	if got.Status != "baseline" {
		t.Fatalf("historical terminal progress should stay baseline: %+v", got)
	}
}

func TestDecideCapturesTerminalSubscriptionFilterConflict(t *testing.T) {
	rep := report{
		Subscription: subscriptionSummary{TerminalFilterMatched: true},
		V3: threadSummary{HighlightedThreads: []threadSurface{{
			ThreadID: "thread-1", AssignmentState: "completed",
		}}},
	}
	got := decide(rep)
	if got.Status != "captured_terminal_live_conflict" {
		t.Fatalf("terminal subscription filter conflict not captured: %+v", got)
	}
}

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
