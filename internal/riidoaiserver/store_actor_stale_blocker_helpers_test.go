package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func assignAndStartStaleBlocker(t *testing.T, store *Store, ctx context.Context) (Assignment, PollResponse) {
	t.Helper()
	first := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-old", RuntimeProvider: "codex", Prompt: "first",
	})
	firstPoll := mustPollActor(t, store, ctx, "agent-old")
	if firstPoll.Action != PollStart || firstPoll.Assignment == nil || firstPoll.Assignment.ID != first.ID {
		t.Fatalf("first poll = %+v", firstPoll)
	}
	return first, firstPoll
}

func assertStaleBlockedAssignmentRepaired(t *testing.T, store *Store, ctx context.Context, second Assignment) {
	t.Helper()
	repairedPoll := mustPollActor(t, store, ctx, "agent-new")
	if repairedPoll.Action != PollStart ||
		repairedPoll.Assignment == nil ||
		repairedPoll.Assignment.ID != second.ID ||
		repairedPoll.Assignment.BlockedByAssignmentID != "" {
		t.Fatalf("new assignment should start after stale blocker repair: %+v", repairedPoll)
	}
}

func assertStaleBlockerRepairOperations(t *testing.T, records []AssignmentOperationRecord, firstID, secondID string) {
	t.Helper()
	var sawRepairEvent, sawFailedBlocker, sawLeasedCurrent bool
	for _, record := range records {
		sawFailedBlocker = sawFailedBlocker || record.Assignment.ID == firstID && record.Assignment.State == AssignmentFailed
		sawLeasedCurrent = sawLeasedCurrent ||
			record.Assignment.ID == secondID &&
				record.Assignment.State == AssignmentLeased &&
				record.Assignment.BlockedByAssignmentID == ""
		for _, event := range record.Events {
			sawRepairEvent = sawRepairEvent ||
				event.AssignmentID == firstID &&
					event.Type == EventAssignmentFailed &&
					strings.Contains(event.Message, "blocked queued assignment")
		}
	}
	if !sawRepairEvent || !sawFailedBlocker || !sawLeasedCurrent {
		t.Fatalf("stale blocker repair evidence missing: %+v", records)
	}
}
