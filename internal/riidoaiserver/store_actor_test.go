package riidoaiserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestStoreActorAssignmentLifecycle(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:      "component-a",
		AgentID:          "agent-1",
		RuntimeProvider:  "codex",
		Prompt:           "ship it",
		AgentInstruction: "act as a release captain",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if assignment.State != AssignmentQueued || assignment.ID != "asn-000001" {
		t.Fatalf("assignment = %+v", assignment)
	}

	now = now.Add(time.Minute)
	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent start: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.State != AssignmentLeased {
		t.Fatalf("poll start = %+v", poll)
	}
	if poll.Assignment.LeaseToken == "" {
		t.Fatalf("poll start lease_token is empty")
	}
	if poll.Assignment.AgentInstruction != "act as a release captain" {
		t.Fatalf("agent_instruction = %q", poll.Assignment.AgentInstruction)
	}

	now = now.Add(time.Minute)
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

	now = now.Add(time.Minute)
	running, err := store.RecordAgentEvent(ctx, "agent-1", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentRunning,
		EventType:    EventAssignmentRunning,
		Message:      "running",
	})
	if err != nil {
		t.Fatalf("RecordAgentEvent running: %v", err)
	}
	if running.Assignment == nil || running.Assignment.State != AssignmentRunning || running.Event.Seq != 3 {
		t.Fatalf("running event = %+v", running)
	}

	now = now.Add(time.Minute)
	completed, err := store.RecordAgentEvent(ctx, "agent-1", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentCompleted,
		EventType:    EventAssignmentCompleted,
		Message:      "done",
	})
	if err != nil {
		t.Fatalf("RecordAgentEvent completed: %v", err)
	}
	if completed.Assignment == nil || completed.Assignment.State != AssignmentCompleted {
		t.Fatalf("completed event = %+v", completed)
	}

	metrics, err := store.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if metrics.TasksTotal != 1 || metrics.AssignmentsTotal != 1 || metrics.AssignmentsByState[AssignmentCompleted] != 1 {
		t.Fatalf("metrics assignments = %+v", metrics)
	}
	if metrics.PollRequestsTotal != 1 || metrics.PollActionsTotal[PollStart] != 1 {
		t.Fatalf("metrics poll = %+v", metrics)
	}
	if metrics.AgentEventsTotal != 2 || metrics.TaskEventsTotal != 4 {
		t.Fatalf("metrics events = %+v", metrics)
	}
}

func TestStoreActorRejectsLongAgentInstruction(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	_, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:      "component-a",
		AgentID:          "agent-1",
		RuntimeProvider:  "codex",
		Prompt:           "ship it",
		AgentInstruction: strings.Repeat("지", AgentInstructionMaxCharacters+1),
	})
	if err == nil || !strings.Contains(err.Error(), "agent_instruction") {
		t.Fatalf("expected agent_instruction validation error, got %v", err)
	}
}

func TestStoreActorReassignmentCancelsPreviousAndBlocksNewAgent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 27, 11, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	first, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "first",
	})
	if err != nil {
		t.Fatalf("AssignTask first: %v", err)
	}
	if _, err := store.PollAgent(ctx, "agent-1", daemonPollRequest()); err != nil {
		t.Fatalf("PollAgent first: %v", err)
	}

	now = now.Add(time.Minute)
	second, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-2",
		RuntimeProvider: "codex",
		Prompt:          "second",
	})
	if err != nil {
		t.Fatalf("AssignTask second: %v", err)
	}
	if second.ReplacesAssignmentID != first.ID || second.BlockedByAssignmentID != first.ID || second.State != AssignmentQueued {
		t.Fatalf("second assignment = %+v", second)
	}

	pollSecond, err := store.PollAgent(ctx, "agent-2", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent second before cancel: %v", err)
	}
	if pollSecond.Action != PollNone {
		t.Fatalf("poll second before cancel = %+v", pollSecond)
	}

	pollFirst, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent first cancel: %v", err)
	}
	if pollFirst.Action != PollCancel || pollFirst.Assignment == nil || pollFirst.Assignment.ID != first.ID {
		t.Fatalf("poll first cancel = %+v", pollFirst)
	}

	if _, err := store.RecordAgentEvent(ctx, "agent-1", AgentEventRequest{
		AssignmentID: first.ID,
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentCancelled,
		EventType:    EventAssignmentCancelled,
	}); err != nil {
		t.Fatalf("RecordAgentEvent cancel: %v", err)
	}

	pollSecond, err = store.PollAgent(ctx, "agent-2", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent second start: %v", err)
	}
	if pollSecond.Action != PollStart || pollSecond.Assignment == nil || pollSecond.Assignment.ID != second.ID {
		t.Fatalf("poll second start = %+v", pollSecond)
	}
}

func TestStoreActorRejectsStoreUnsafeProviderStatus(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	if _, err := store.SyncProviderStatus(ctx, "agent-1", ProviderStatusSyncRequest{
		DaemonID:            "daemon-1",
		RuntimeID:           "runtime-1",
		DistributionChannel: hostintegration.DistributionChannelDevLocal,
		Providers: []ProviderStatusRecord{{
			ProviderKind:  capability.ProviderKind("codex"),
			RoutingStatus: hostintegration.ProviderRoutingLoginRequired,
		}},
	}); err != nil {
		t.Fatalf("SyncProviderStatus: %v", err)
	}

	if _, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "blocked",
	}); err == nil || !strings.Contains(err.Error(), "provider codex cannot be assigned") {
		t.Fatalf("AssignTask blocked err=%v", err)
	}
}

func TestStoreActorValidatesAgentRuntimeBinding(t *testing.T) {
	ctx := context.Background()
	registry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{{
		AgentID:         "agent-1",
		DaemonID:        "daemon-1",
		RuntimeID:       "runtime-1",
		RuntimeProvider: "codex",
	}})
	if err != nil {
		t.Fatalf("NewStaticAgentRegistry: %v", err)
	}
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: registry})
	defer store.Close()

	if _, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "claude",
		Prompt:          "wrong provider",
	}); err == nil || !strings.Contains(err.Error(), "bound to runtime_provider codex") {
		t.Fatalf("AssignTask wrong provider err=%v", err)
	}

	if _, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "right provider",
	}); err != nil {
		t.Fatalf("AssignTask right provider: %v", err)
	}
	if _, err := store.PollAgent(ctx, "agent-1", PollRequest{DaemonID: "other", RuntimeID: "runtime-1"}); err == nil || !strings.Contains(err.Error(), "bound to daemon_id daemon-1") {
		t.Fatalf("PollAgent wrong daemon err=%v", err)
	}
}

func daemonPollRequest() PollRequest {
	return PollRequest{DaemonID: "daemon-1", RuntimeID: "runtime-1"}
}
