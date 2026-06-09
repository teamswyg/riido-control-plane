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

	now = now.Add(5 * time.Second)
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

	now = now.Add(5 * time.Second)
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

func TestStoreActorPersistsProviderSessionID(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 9, 11, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		ModelID:         "codex-default",
		Prompt:          "ship it",
		ResumeSessionID: "th-prev",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if assignment.ResumeSessionID != "th-prev" {
		t.Fatalf("resume_session_id = %q, want %q", assignment.ResumeSessionID, "th-prev")
	}

	now = now.Add(time.Second)
	pinned, err := store.RecordAgentEvent(ctx, "agent-1", AgentEventRequest{
		AssignmentID:      assignment.ID,
		TaskID:            assignment.TaskID,
		DaemonID:          "daemon-1",
		RuntimeID:         "runtime-1",
		EventType:         EventProviderSessionPinned,
		ProviderSessionID: "th-current",
	})
	if err != nil {
		t.Fatalf("RecordAgentEvent session pinned: %v", err)
	}
	if pinned.Assignment == nil || pinned.Assignment.ProviderSessionID != "th-current" {
		t.Fatalf("provider session assignment = %+v", pinned.Assignment)
	}
	if got := pinned.Event.Metadata[assignmentMetadataProviderSessionID]; got != "th-current" {
		t.Fatalf("provider session event metadata = %+v", pinned.Event.Metadata)
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

	now = now.Add(5 * time.Second)
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

func TestStoreActorReassignmentCancelsQueuedPreviousWithoutBlockingNewAgent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 5, 11, 0, 0, 0, time.UTC)
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

	now = now.Add(time.Second)
	second, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-2",
		RuntimeProvider: "codex",
		Prompt:          "second",
	})
	if err != nil {
		t.Fatalf("AssignTask second: %v", err)
	}
	if second.ReplacesAssignmentID != first.ID || second.BlockedByAssignmentID != "" {
		t.Fatalf("second assignment should replace without blocker: %+v", second)
	}
	firstProjection, ok, err := store.LoadAssignmentProjection(ctx, first.ID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection first: %v", err)
	}
	if !ok || firstProjection.Assignment.State != AssignmentCancelled {
		t.Fatalf("first projection = %+v ok=%v", firstProjection, ok)
	}
	secondPoll, err := store.PollAgent(ctx, "agent-2", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent second: %v", err)
	}
	if secondPoll.Action != PollStart || secondPoll.Assignment == nil || secondPoll.Assignment.ID != second.ID {
		t.Fatalf("second poll = %+v", secondPoll)
	}
	firstPoll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent first: %v", err)
	}
	if firstPoll.Action != PollNone {
		t.Fatalf("first poll after queued cancellation = %+v", firstPoll)
	}
}

func TestStoreActorClientStopCancelsActiveAssignmentForDaemonPoll(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 5, 13, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "first",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	pollStart, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent start: %v", err)
	}
	if pollStart.Action != PollStart || pollStart.Assignment == nil || pollStart.Assignment.ID != assignment.ID {
		t.Fatalf("poll start = %+v", pollStart)
	}

	now = now.Add(time.Second)
	cancelled, err := store.CancelAssignment(ctx, "task-a", CancelAssignmentRequest{
		AgentID:      "agent-1",
		AssignmentID: assignment.ID,
		Reason:       "user requested stop",
	})
	if err != nil {
		t.Fatalf("CancelAssignment: %v", err)
	}
	if cancelled.State != AssignmentCancelling {
		t.Fatalf("cancelled assignment = %+v", cancelled)
	}
	projection, ok, err := store.LoadAssignmentProjection(ctx, assignment.ID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection: %v", err)
	}
	if !ok || projection.Assignment.State != AssignmentCancelling {
		t.Fatalf("projection after cancel = %+v ok=%v", projection, ok)
	}

	pollCancel, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent cancel: %v", err)
	}
	if pollCancel.Action != PollCancel || pollCancel.Assignment == nil || pollCancel.Assignment.ID != assignment.ID {
		t.Fatalf("poll cancel = %+v", pollCancel)
	}
}

func TestStoreActorClientStopCancelsQueuedAssignmentImmediately(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 5, 14, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "first",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	cancelled, err := store.CancelAssignment(ctx, "task-a", CancelAssignmentRequest{
		AgentID:      "agent-1",
		AssignmentID: assignment.ID,
		Reason:       "user requested stop",
	})
	if err != nil {
		t.Fatalf("CancelAssignment: %v", err)
	}
	if cancelled.State != AssignmentCancelled {
		t.Fatalf("cancelled queued assignment = %+v", cancelled)
	}
	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Action != PollNone {
		t.Fatalf("poll after queued cancel = %+v", poll)
	}
}

func TestStoreActorReassigningSameBlockedQueuedAssignmentRepairsQueuedBlocker(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	blocker := replayOperationRecord("task-a", "asn-000020", "agent-old", AssignmentQueued, 1, base)
	current := replayOperationRecord("task-a", "asn-000027", "agent-new", AssignmentQueued, 2, base.Add(time.Second))
	current.Assignment.BlockedByAssignmentID = blocker.Assignment.ID
	current.Assignment.ReplacesAssignmentID = blocker.Assignment.ID
	current.OperationID = assignmentOperationID(current.OperationType, current.Assignment, current.Events)
	operations := &runtimeFakeAssignmentOperationStore{records: []AssignmentOperationRecord{blocker, current}}
	now := base.Add(2 * time.Second)
	store, err := OpenStoreWithConfig(ctx, StoreConfig{
		Now:            func() time.Time { return now },
		OperationStore: operations,
	})
	if err != nil {
		t.Fatalf("OpenStoreWithConfig: %v", err)
	}
	defer store.Close()

	reassigned, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-new",
		RuntimeProvider: "codex",
		Prompt:          "second",
	})
	if err != nil {
		t.Fatalf("AssignTask same blocked current: %v", err)
	}
	if reassigned.ID != current.Assignment.ID || reassigned.BlockedByAssignmentID != "" {
		t.Fatalf("reassigned current = %+v", reassigned)
	}
	blockerProjection, ok, err := store.LoadAssignmentProjection(ctx, blocker.Assignment.ID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection blocker: %v", err)
	}
	if !ok || blockerProjection.Assignment.State != AssignmentCancelled {
		t.Fatalf("blocker projection = %+v ok=%v", blockerProjection, ok)
	}
	currentProjection, ok, err := store.LoadAssignmentProjection(ctx, current.Assignment.ID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection current: %v", err)
	}
	if !ok || currentProjection.Assignment.BlockedByAssignmentID != "" {
		t.Fatalf("current projection = %+v ok=%v", currentProjection, ok)
	}
	poll, err := store.PollAgent(ctx, "agent-new", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent new: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.ID != current.Assignment.ID {
		t.Fatalf("poll after blocker repair = %+v", poll)
	}
}

func TestStoreActorPollRepairsStaleBlockedQueuedAssignment(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 9, 4, 0, 0, 0, time.UTC)
	operations := &runtimeFakeActiveLeaseOperationStore{}
	store := NewStoreWithConfig(StoreConfig{
		Now:                 func() time.Time { return now },
		ActiveLeaseDuration: time.Minute,
		OperationStore:      operations,
	})
	defer store.Close()

	first, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-old",
		RuntimeProvider: "codex",
		Prompt:          "first",
	})
	if err != nil {
		t.Fatalf("AssignTask first: %v", err)
	}
	firstPoll, err := store.PollAgent(ctx, "agent-old", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent first: %v", err)
	}
	if firstPoll.Action != PollStart || firstPoll.Assignment == nil || firstPoll.Assignment.ID != first.ID {
		t.Fatalf("first poll = %+v", firstPoll)
	}
	operations.activeFound = true
	operations.activeLease = AssignmentActiveLease{
		AgentID:            "agent-old",
		ActiveAssignmentID: first.ID,
		LeaseToken:         firstPoll.Assignment.LeaseToken,
		HeartbeatAt:        now,
		LeaseExpiresAt:     now.Add(time.Minute),
		LeaseExpiresUnixMS: now.Add(time.Minute).UnixMilli(),
	}

	now = now.Add(10 * time.Second)
	second, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-new",
		RuntimeProvider: "codex",
		Prompt:          "second",
	})
	if err != nil {
		t.Fatalf("AssignTask second: %v", err)
	}
	if second.ReplacesAssignmentID != first.ID || second.BlockedByAssignmentID != first.ID {
		t.Fatalf("second assignment = %+v", second)
	}

	beforeExpiry, err := store.PollAgent(ctx, "agent-new", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent new before expiry: %v", err)
	}
	if beforeExpiry.Action != PollNone {
		t.Fatalf("new assignment should stay blocked before lease expiry: %+v", beforeExpiry)
	}

	now = now.Add(2 * time.Minute)
	repairedPoll, err := store.PollAgent(ctx, "agent-new", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent new after stale blocker: %v", err)
	}
	if repairedPoll.Action != PollStart ||
		repairedPoll.Assignment == nil ||
		repairedPoll.Assignment.ID != second.ID ||
		repairedPoll.Assignment.BlockedByAssignmentID != "" {
		t.Fatalf("new assignment should start after stale blocker repair: %+v", repairedPoll)
	}

	oldPollAfterRepair, err := store.PollAgent(ctx, "agent-old", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent old after repair: %v", err)
	}
	if oldPollAfterRepair.Action != PollNone {
		t.Fatalf("old assignment should not stay active after stale blocker repair: %+v", oldPollAfterRepair)
	}

	var sawRepairEvent bool
	var sawFailedBlocker bool
	var sawLeasedCurrent bool
	for _, record := range operations.records {
		if record.Assignment.ID == first.ID && record.Assignment.State == AssignmentFailed {
			sawFailedBlocker = true
		}
		if record.Assignment.ID == second.ID &&
			record.Assignment.State == AssignmentLeased &&
			record.Assignment.BlockedByAssignmentID == "" {
			sawLeasedCurrent = true
		}
		for _, event := range record.Events {
			if event.AssignmentID == first.ID &&
				event.Type == EventAssignmentFailed &&
				strings.Contains(event.Message, "blocked queued assignment") {
				sawRepairEvent = true
			}
		}
	}
	if !sawRepairEvent {
		t.Fatalf("missing stale blocker repair event: %+v", operations.records)
	}
	if !sawFailedBlocker {
		t.Fatalf("missing failed blocker operation: %+v", operations.records)
	}
	if !sawLeasedCurrent {
		t.Fatalf("missing leased current operation with cleared blocker: %+v", operations.records)
	}
}

func TestStoreActorAdditiveAssignmentKeepsExistingAgentActive(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	first, err := store.AssignTaskAdditive(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "first",
	})
	if err != nil {
		t.Fatalf("AssignTaskAdditive first: %v", err)
	}
	firstPoll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent first: %v", err)
	}
	if firstPoll.Action != PollStart || firstPoll.Assignment == nil || firstPoll.Assignment.ID != first.ID {
		t.Fatalf("first poll = %+v", firstPoll)
	}

	now = now.Add(5 * time.Second)
	second, err := store.AssignTaskAdditive(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-2",
		RuntimeProvider: "codex",
		Prompt:          "second",
	})
	if err != nil {
		t.Fatalf("AssignTaskAdditive second: %v", err)
	}
	if second.ReplacesAssignmentID != "" || second.BlockedByAssignmentID != "" {
		t.Fatalf("additive assignment must not replace/block existing assignment: %+v", second)
	}

	firstPoll, err = store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent first active: %v", err)
	}
	if firstPoll.Action != PollActive || firstPoll.Assignment == nil || firstPoll.Assignment.ID != first.ID {
		t.Fatalf("first active poll = %+v", firstPoll)
	}
	secondPoll, err := store.PollAgent(ctx, "agent-2", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent second: %v", err)
	}
	if secondPoll.Action != PollStart || secondPoll.Assignment == nil || secondPoll.Assignment.ID != second.ID {
		t.Fatalf("second poll = %+v", secondPoll)
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
