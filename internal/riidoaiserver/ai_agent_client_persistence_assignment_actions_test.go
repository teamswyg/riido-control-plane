package riidoaiserver

import (
	"context"
	"testing"
)

func TestPersistentAIAgentClientStorePersistsAgentAssignmentStop(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	assigned, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-persist-stop", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-persist-stop",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment: %v", err)
	}
	stopped, err := store.StopAIAgentTaskAgentAssignment(ctx, principal, assigned.TaskID, assigned.AgentID,
		AgentAssignmentActionRequest{AssignmentID: assigned.AssignmentID, Reason: "user stop"})
	if err != nil {
		t.Fatalf("StopAIAgentTaskAgentAssignment: %v", err)
	}
	if stopped.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("stop response = %+v", stopped)
	}

	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen persistent store: %v", err)
	}
	threads, err := reopened.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if threads.ActiveStream != nil {
		t.Fatalf("stopped assignment restored active stream: %+v", threads.ActiveStream)
	}
	thread := actionThreadByID(t, threads, assigned.ThreadID)
	if thread.AssignmentState != AgentAssignmentStateStopped ||
		thread.CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("restored stopped thread = %+v", thread)
	}
}

func actionThreadByID(t *testing.T, threads AIAgentTaskThreadCollectionResponse, threadID string) AIAgentTaskThreadRecord {
	t.Helper()
	for _, thread := range threads.Threads {
		if thread.ThreadID == threadID {
			return thread
		}
	}
	t.Fatalf("thread %q not found in %+v", threadID, threads.Threads)
	return AIAgentTaskThreadRecord{}
}
