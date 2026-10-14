package riidoaiserver

import "testing"

func TestQueuedClientEventFilterRetainsUnresolvedEvents(t *testing.T) {
	store := queuedFilterStore()
	queued := queuedFilterStatusEvent(1, "task", "missing", AgentTaskCommentQueuedByBusyAgent)
	if queuedClientEventIsSupersededLocked(store, queued, nil) {
		t.Fatalf("missing thread must not hide queued event")
	}
	nonQueued := queuedFilterStatusEvent(2, "task", "missing", AgentTaskCommentRuntimeProgress)
	if queuedClientEventIsSupersededLocked(store, nonQueued, nil) {
		t.Fatalf("non-queued status must not be handled by queued filter")
	}
}

func TestEventConversationIDLockedRequiresThreadReference(t *testing.T) {
	store := queuedFilterStore()
	if got := eventConversationIDLocked(store, ClientStreamEvent{Payload: "unknown"}); got != "" {
		t.Fatalf("unknown payload conversation = %q, want empty", got)
	}
	event := queuedFilterStatusEvent(1, "task", "missing", AgentTaskCommentRuntimeProgress)
	if got := eventConversationIDLocked(store, event); got != "" {
		t.Fatalf("missing thread conversation = %q, want empty", got)
	}
}
