package main

import "testing"

func TestCandidateConversationsRanksActiveThreads(t *testing.T) {
	threads := threadCollection{Threads: []threadRecord{
		{ThreadID: "terminal", ConversationID: "conv-a", AssignmentState: "completed"},
		{ThreadID: "active", ConversationID: "conv-b", WorkStatus: "running"},
		{ThreadID: "queued", ConversationID: "conv-c", AssignmentState: "queued"},
	}}
	got, count := candidateConversations(threads, threadCollection{}, 25)
	if count != 3 || len(got) != 3 {
		t.Fatalf("unexpected candidates count=%d got=%d", count, len(got))
	}
	if got[0].ConversationID != "conv-b" || got[0].RunningCount != 1 {
		t.Fatalf("running conversation should rank first: %+v", got)
	}
	if got[1].ConversationID != "conv-c" || got[1].QueuedCount != 1 {
		t.Fatalf("queued conversation should rank second: %+v", got)
	}
}

func TestCandidateConversationsFallsBackToV2AndThreadID(t *testing.T) {
	v2 := threadCollection{Threads: []threadRecord{{
		ThreadID: "thread-only", AssignmentID: "asn-1", RunID: "run-1",
		WorkStatus: "running",
	}}}
	got, count := candidateConversations(threadCollection{}, v2, 25)
	if count != 1 || len(got) != 1 {
		t.Fatalf("unexpected candidates count=%d got=%d", count, len(got))
	}
	candidate := got[0]
	if candidate.ConversationID != "thread-only" {
		t.Fatalf("conversation fallback = %q", candidate.ConversationID)
	}
	if candidate.AssignmentID != "asn-1" || candidate.RunID != "run-1" {
		t.Fatalf("missing sample IDs: %+v", candidate)
	}
}
