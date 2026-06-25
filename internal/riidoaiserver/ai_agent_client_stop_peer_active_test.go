package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentClientStopInactiveAgentPreservesPeerActiveStream(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}

	done, err := store.AssignAIAgentTask(ctx, principal, "task-stop-peer", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-peer-done",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask done: %v", err)
	}
	recordAssignmentCompleted(t, store, done, "done")

	active, err := store.AssignAIAgentTask(ctx, principal, "task-stop-peer", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-peer-active",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask active: %v", err)
	}

	_, err = store.StopAIAgentTaskAgentAssignment(
		ctx,
		principal,
		done.TaskID,
		done.AgentID,
		AgentAssignmentActionRequest{Reason: "user stop"},
	)
	if err == nil {
		t.Fatal("StopAIAgentTaskAgentAssignment succeeded for inactive agent")
	}

	threads, err := store.ListAIAgentTaskThreads(ctx, principal, done.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if threads.ActiveStream == nil || threads.ActiveStream.ThreadID != active.ThreadID {
		t.Fatalf("peer active_stream was not preserved: %+v", threads.ActiveStream)
	}
	assertStopPeerThreadState(t, threads.Threads, done.AssignmentID, AgentAssignmentStateCompleted)
	assertStopPeerThreadState(t, threads.Threads, active.AssignmentID, active.AssignmentState)
}

func assertStopPeerThreadState(
	t *testing.T,
	threads []AIAgentTaskThreadRecord,
	assignmentID string,
	want AgentAssignmentState,
) {
	t.Helper()
	for _, thread := range threads {
		if thread.AssignmentID != assignmentID {
			continue
		}
		if thread.AssignmentState != want {
			t.Fatalf("thread %s state = %s, want %s", assignmentID, thread.AssignmentState, want)
		}
		return
	}
	t.Fatalf("missing thread for assignment %s: %+v", assignmentID, threads)
}
