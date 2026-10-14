package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentClientStopWithoutActiveThreadDoesNotCreateSyntheticThread(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}

	_, err := store.StopAIAgentTaskAgentAssignment(
		ctx,
		principal,
		"task-stop-no-active",
		"agent-public-openclaw",
		AgentAssignmentActionRequest{Reason: "user stop"},
	)
	if err == nil {
		t.Fatal("StopAIAgentTaskAgentAssignment succeeded without an active thread")
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, "task-stop-no-active")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 0 {
		t.Fatalf("stop without active work created synthetic threads: %+v", threads.Threads)
	}
}

func TestAIAgentClientStopWithoutActiveThreadPreservesCompletedThread(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-stop-completed", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-completed-stop",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	recordAssignmentCompleted(t, store, assigned, "done")

	_, err = store.StopAIAgentTaskAgentAssignment(
		ctx,
		principal,
		assigned.TaskID,
		assigned.AgentID,
		AgentAssignmentActionRequest{Reason: "user stop"},
	)
	if err == nil {
		t.Fatal("StopAIAgentTaskAgentAssignment succeeded after completion")
	}
	thread := lastThread(t, store, assigned.TaskID)
	if thread.AssignmentState != AgentAssignmentStateCompleted ||
		thread.CommentKind != AgentTaskCommentTaskCompleted ||
		thread.ResultMessage != "done" {
		t.Fatalf("stop without active work mutated completed history: %+v", thread)
	}
}
