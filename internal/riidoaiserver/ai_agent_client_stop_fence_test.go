package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

// C2: once the user stops a task, a late runtime-progress event that raced the
// stop must not re-activate the thread (flip it back to running, re-open the SSE
// active_stream, re-arm the agent, or append more lines). Both ingestion paths
// — /events riido_log and the /thread-progress batch — must drop it.
func TestC2LateRuntimeProgressDoesNotReactivateStoppedThread(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}

	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-c2", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}

	log := func(msg string) TaskEvent {
		return TaskEvent{
			TaskID:       "task-c2",
			AssignmentID: "asn-c2",
			AgentID:      assigned.AgentID,
			Type:         EventRiidoLog,
			State:        AssignmentRunning,
			Message:      msg,
			At:           time.Now().UTC(),
		}
	}

	// Advance the thread to running via a runtime-progress log.
	if err := store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, log("thinking")); err != nil {
		t.Fatalf("running log: %v", err)
	}
	if th := lastThread(t, store, "task-c2"); !taskThreadHasActiveStream(th) {
		t.Fatalf("thread should be active before stop: %+v", th)
	}

	// User stops the task.
	if _, err := store.StopAIAgentTask(ctx, principal, "task-c2", StopAIAgentTaskRequest{}); err != nil {
		t.Fatalf("StopAIAgentTask: %v", err)
	}
	stopped := lastThread(t, store, "task-c2")
	if stopped.AssignmentState != AgentAssignmentStateStopped || taskThreadHasActiveStream(stopped) {
		t.Fatalf("thread should be stopped after stop: %+v", stopped)
	}
	linesAfterStop := len(stopped.Lines)

	// A late /events riido_log races the stop. It must be dropped.
	if err := store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, log("therefore the answer is")); err != nil {
		t.Fatalf("late log: %v", err)
	}
	after := lastThread(t, store, "task-c2")
	if after.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("late riido_log re-activated stopped thread: state=%s", after.AssignmentState)
	}
	if taskThreadHasActiveStream(after) {
		t.Fatal("late riido_log re-opened active_stream on a stopped thread")
	}
	if after.WorkStatus == AgentWorkStatusRunning {
		t.Fatal("late riido_log re-armed work status to running")
	}
	if len(after.Lines) != linesAfterStop {
		t.Fatalf("late riido_log appended a line to a stopped thread: %d -> %d", linesAfterStop, len(after.Lines))
	}

	// The /thread-progress batch path is fenced too: 0 accepted, no reactivation.
	resp, err := store.RecordAIAgentThreadProgress(ctx, assigned.AgentID, AgentThreadProgressBatchRequest{
		TaskID:       "task-c2",
		AssignmentID: "asn-c2",
		ThreadID:     after.ThreadID,
		RunID:        after.RunID,
		Lines:        []AgentThreadProgressLine{{Seq: 99, Message: "late batch"}},
	})
	if err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
	if resp.AcceptedLines != 0 {
		t.Fatalf("late batch accepted %d lines, want 0 (fenced)", resp.AcceptedLines)
	}
	if final := lastThread(t, store, "task-c2"); taskThreadHasActiveStream(final) || final.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("late batch re-activated stopped thread: %+v", final)
	}
}

// Regression: a genuinely active (running) thread still advances on runtime
// progress — the fence only blocks stopped/terminal threads.
func TestC2RunningThreadStillAcceptsRuntimeProgress(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-c2-ok", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	if err := store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, TaskEvent{
		TaskID:       "task-c2-ok",
		AssignmentID: "asn-ok",
		AgentID:      assigned.AgentID,
		Type:         EventRiidoLog,
		State:        AssignmentRunning,
		Message:      "working",
		At:           time.Now().UTC(),
	}); err != nil {
		t.Fatalf("running log: %v", err)
	}
	th := lastThread(t, store, "task-c2-ok")
	if !taskThreadHasActiveStream(th) || th.AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("running thread should accept progress and stay active: %+v", th)
	}
	if len(th.Lines) == 0 {
		t.Fatal("running progress line was not recorded")
	}
}

func TestUnassignTerminalFailedThreadUpdatesOriginalThread(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-terminal-unassign", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-terminal-unassign",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	if err := store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, TaskEvent{
		TaskID:       assigned.TaskID,
		AssignmentID: assigned.AssignmentID,
		AgentID:      assigned.AgentID,
		Type:         EventAssignmentFailed,
		State:        AssignmentFailed,
		Message:      "agent work failed",
		At:           time.Now().UTC(),
	}); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent failed: %v", err)
	}
	failed := lastThread(t, store, assigned.TaskID)
	if failed.ThreadID != assigned.ThreadID ||
		failed.AssignmentID != assigned.AssignmentID ||
		failed.AssignmentState != AgentAssignmentStateFailed {
		t.Fatalf("test setup should fail original thread: assigned=%+v failed=%+v", assigned, failed)
	}

	unassigned, err := store.UnassignAIAgentTask(ctx, principal, assigned.TaskID, UnassignAIAgentTaskRequest{
		AgentID: assigned.AgentID,
		Reason:  "participant_removed",
	})
	if err != nil {
		t.Fatalf("UnassignAIAgentTask: %v", err)
	}
	if unassigned.ThreadID != assigned.ThreadID ||
		unassigned.AssignmentID != assigned.AssignmentID ||
		unassigned.RunID != assigned.RunID {
		t.Fatalf("unassign should target original terminal thread: assigned=%+v unassigned=%+v", assigned, unassigned)
	}

	threads, err := store.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if threads.ActiveStream != nil {
		t.Fatalf("terminal unassign must not expose active_stream: %+v", threads.ActiveStream)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("terminal unassign should update in place, not append synthetic rows: %+v", threads.Threads)
	}
	thread := threads.Threads[0]
	if thread.ThreadID != assigned.ThreadID ||
		thread.AssignmentID != assigned.AssignmentID ||
		thread.AssignmentState != AgentAssignmentStateStopped ||
		thread.WorkStatus != AgentWorkStatusIdle {
		t.Fatalf("terminal unassign did not stop original thread: %+v", thread)
	}
}

func TestStoppedThreadIgnoresLateProviderLogAndCancellingReplay(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-late-provider-log", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-late-provider-log",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	stopped, err := store.StopAIAgentTask(ctx, principal, assigned.TaskID, StopAIAgentTaskRequest{
		AgentID: assigned.AgentID,
		Reason:  "participant_removed",
	})
	if err != nil {
		t.Fatalf("StopAIAgentTask: %v", err)
	}
	if stopped.ThreadID != assigned.ThreadID || stopped.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("stopped response = %+v, assigned=%+v", stopped, assigned)
	}

	if err := store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, TaskEvent{
		TaskID:       assigned.TaskID,
		AssignmentID: assigned.AssignmentID,
		AgentID:      assigned.AgentID,
		Type:         EventProviderLog,
		State:        AssignmentCancelling,
		Message:      "codex rate limits updated",
		At:           time.Now().UTC(),
	}); err != nil {
		t.Fatalf("late provider log: %v", err)
	}

	threads, err := store.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if threads.ActiveStream != nil {
		t.Fatalf("late provider log must not reopen active_stream: %+v", threads.ActiveStream)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %+v", threads.Threads)
	}
	thread := threads.Threads[0]
	if thread.AssignmentState != AgentAssignmentStateStopped ||
		thread.WorkStatus != AgentWorkStatusIdle ||
		thread.Message != stopped.Message {
		t.Fatalf("late provider log reactivated or rewrote stopped thread: stopped=%+v thread=%+v", stopped, thread)
	}
}

func lastThread(t *testing.T, store *DevelopmentAIAgentClientStore, taskID string) AIAgentTaskThreadRecord {
	t.Helper()
	threads := store.taskThreads[taskID]
	if len(threads) == 0 {
		t.Fatalf("no task thread for %s", taskID)
	}
	return threads[len(threads)-1]
}
