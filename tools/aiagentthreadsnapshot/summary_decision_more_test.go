package main

import (
	"encoding/json"
	"testing"
)

func TestDecideSeparatesEndpointErrorAndLiveProgress(t *testing.T) {
	t.Parallel()
	partial := decide(report{Endpoints: []endpointObservation{{Name: "v3"}}})
	if partial.Status != "partial" {
		t.Fatalf("endpoint status 0 should be partial: %+v", partial)
	}
	captured := decide(report{SSEEvents: []sseEventSummary{{WorkStatus: "running"}}})
	if captured.Status != "captured" {
		t.Fatalf("live event should be captured: %+v", captured)
	}
}

func TestSummarizeThreadsCountsAndHighlights(t *testing.T) {
	t.Parallel()
	payload := threadCollection{
		ActiveStream: json.RawMessage(`{"href":"/events"}`),
		Threads: []threadRecord{
			{ThreadID: "root", ConversationID: "conv", WorkStatus: "queued"},
			{
				ThreadID: "child", ConversationID: "other", AssignmentState: "completed",
				ActiveStream: json.RawMessage(`{"href":"/events"}`),
				Messages:     []messageRecord{{Role: "agent"}},
				Lines:        []lineRecord{{Seq: 1}},
			},
		},
	}
	got := summarizeThreads(payload, "child")
	if !got.ActiveStream || got.QueuedCount != 1 || got.TerminalActiveStreamCount != 1 {
		t.Fatalf("unexpected thread counts: %+v", got)
	}
	if len(got.HighlightedThreads) != 1 || got.HighlightedThreads[0].ThreadID != "child" {
		t.Fatalf("thread id should act as highlight fallback: %+v", got.HighlightedThreads)
	}
}

func TestMatchesLiveEventSkipsEmptyAndMatchesRun(t *testing.T) {
	t.Parallel()
	thread := threadSurface{RunID: "run-1"}
	events := []sseEventSummary{
		{RunID: "run-1"},
		{RunID: "run-1", AssignmentState: "running"},
	}
	if !matchesLiveEvent(thread, events) {
		t.Fatal("expected non-empty live event to match run id")
	}
	if matchesLiveEvent(thread, []sseEventSummary{{RunID: "run-1"}}) {
		t.Fatal("empty live event should be ignored")
	}
}
