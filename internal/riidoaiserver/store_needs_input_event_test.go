package riidoaiserver

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestNeedsInputCompletesLeaseAndUnblocksNextAssignment(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	first := mustAssignActorTask(t, store, ctx, "task-first", AssignRequest{
		AgentID: "agent-1", RuntimeProvider: "claude", Prompt: "first",
	})
	poll := mustPollActor(t, store, ctx, first.AgentID)
	mustRecordActorEvent(t, store, ctx, first.AgentID, AgentEventRequest{
		AssignmentID: first.ID, DaemonID: "daemon-1", RuntimeID: "runtime-1",
		State: AssignmentRunning, EventType: EventAssignmentRunning,
	})
	second := mustAssignActorTask(t, store, ctx, "task-second", AssignRequest{
		AgentID: "agent-1", RuntimeProvider: "claude", Prompt: "second",
	})
	blocked := mustPollActor(t, store, ctx, first.AgentID)
	if blocked.Action != PollActive || blocked.Assignment == nil || blocked.Assignment.ID != poll.Assignment.ID {
		t.Fatalf("poll while first is running = %+v", blocked)
	}

	result := mustRecordActorEvent(t, store, ctx, first.AgentID, AgentEventRequest{
		AssignmentID: first.ID, DaemonID: "daemon-1", RuntimeID: "runtime-1",
		State: AssignmentRunning, EventType: EventAssignmentStateUpdated,
		Message:  "어떤 작업부터 진행할까요?",
		Metadata: map[string]string{metadatakeys.AssignmentResultStatus.String(): "needs_input"},
	})
	if result.Assignment.State != AssignmentCompleted || result.Event.State != AssignmentCompleted {
		t.Fatalf("needs-input durable state = assignment:%s event:%s", result.Assignment.State, result.Event.State)
	}
	projection := assignmentEventActionResponse(AIAgentTaskThreadRecord{
		TaskID: first.TaskID, AssignmentID: first.ID, AgentID: first.AgentID,
	}, result.Event.State, result.Event.Message, result.Event.Metadata)
	if projection.WorkStatus != AgentWorkStatusWaitingForUser ||
		projection.AssignmentState != AgentAssignmentStateWaiting || projection.ActiveStream != nil {
		t.Fatalf("needs-input client projection = %+v", projection)
	}

	next := mustPollActor(t, store, ctx, first.AgentID)
	if next.Action != PollStart || next.Assignment == nil || next.Assignment.ID != second.ID {
		t.Fatalf("next poll = %+v, want start %s", next, second.ID)
	}
}
