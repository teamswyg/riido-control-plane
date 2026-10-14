package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStoreActorAssignmentLifecycle(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID:      "component-a",
		AgentID:          "agent-1",
		RuntimeProvider:  "codex",
		Prompt:           "ship it",
		AgentInstruction: "act as a release captain",
	})
	if assignment.State != AssignmentQueued || assignment.ID != "asn-000001" {
		t.Fatalf("assignment = %+v", assignment)
	}

	now = now.Add(5 * time.Second)
	poll := mustPollActor(t, store, ctx, "agent-1")
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.State != AssignmentLeased {
		t.Fatalf("poll start = %+v", poll)
	}
	if poll.Assignment.LeaseToken == "" || poll.Assignment.AgentInstruction != "act as a release captain" {
		t.Fatalf("poll assignment = %+v", poll.Assignment)
	}

	now = now.Add(5 * time.Second)
	heartbeat, err := store.HeartbeatAgent(ctx, "agent-1", AgentHeartbeatRequest{
		DaemonID:            "daemon-1",
		RuntimeID:           "runtime-1",
		ActiveAssignmentIDs: []string{poll.Assignment.ID},
	})
	if err != nil {
		t.Fatalf("HeartbeatAgent: %v", err)
	}
	if len(heartbeat.RefreshedAssignments) != 1 || !heartbeat.RefreshedAssignments[0].UpdatedAt.Equal(now) {
		t.Fatalf("heartbeat = %+v", heartbeat)
	}

	now = now.Add(5 * time.Second)
	running := mustRecordActorEvent(t, store, ctx, "agent-1", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentRunning,
		EventType:    EventAssignmentRunning,
		Message:      "running",
	})
	if running.Assignment == nil || running.Assignment.State != AssignmentRunning || running.Event.Seq != 3 {
		t.Fatalf("running event = %+v", running)
	}

	now = now.Add(5 * time.Second)
	completed := mustRecordActorEvent(t, store, ctx, "agent-1", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentCompleted,
		EventType:    EventAssignmentCompleted,
		Message:      "done",
	})
	assertStoreActorLifecycleMetrics(t, store, ctx, completed)
}
