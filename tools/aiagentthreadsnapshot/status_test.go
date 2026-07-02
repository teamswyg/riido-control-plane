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
