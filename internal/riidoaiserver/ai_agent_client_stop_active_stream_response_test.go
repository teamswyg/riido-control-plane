package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAIAgentClientStopResponseTargetsCurrentActiveStream(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	now := time.Now().UTC()

	store.taskThreads["task-stop-visible"] = []AIAgentTaskThreadRecord{
		visibleStopThread("asn-running-old", "run-old", AgentAssignmentStateRunning, now),
		visibleStopThread("asn-queued-new", "run-new", AgentAssignmentStateQueued, now.Add(time.Second)),
	}

	before, err := store.ListAIAgentTaskThreads(ctx, principal, "task-stop-visible")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads before stop: %v", err)
	}
	if before.ActiveStream == nil || before.ActiveStream.ThreadID != "thread-asn-queued-new" {
		t.Fatalf("pre-stop active_stream = %+v", before.ActiveStream)
	}

	stopped, err := store.StopAIAgentTaskAgentAssignment(ctx, principal, "task-stop-visible", "agent-public-openclaw", AgentAssignmentActionRequest{Reason: "user stop"})
	if err != nil {
		t.Fatalf("StopAIAgentTaskAgentAssignment: %v", err)
	}
	if stopped.ThreadID != before.ActiveStream.ThreadID || stopped.AssignmentID != "asn-queued-new" {
		t.Fatalf("stop response should target current active_stream: stopped=%+v active=%+v", stopped, before.ActiveStream)
	}

	after, err := store.ListAIAgentTaskThreads(ctx, principal, "task-stop-visible")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads after stop: %v", err)
	}
	if after.ActiveStream != nil {
		t.Fatalf("stop must close the collection active_stream: %+v", after.ActiveStream)
	}
	for _, thread := range after.Threads {
		if thread.AssignmentState != AgentAssignmentStateStopped || thread.WorkStatus != AgentWorkStatusIdle {
			t.Fatalf("thread should be stopped: %+v", thread)
		}
	}
}

func visibleStopThread(assignmentID, runID string, state AgentAssignmentState, at time.Time) AIAgentTaskThreadRecord {
	return AIAgentTaskThreadRecord{
		ThreadID:        "thread-" + assignmentID,
		TaskID:          "task-stop-visible",
		AssignmentID:    assignmentID,
		AgentID:         "agent-public-openclaw",
		RunID:           runID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: state,
		CommentKind:     AgentTaskCommentAssignmentStarted,
		StartedAt:       at,
	}
}
