package riidoaiserver

import (
	"context"
	"testing"
)

func TestAdditiveRequestIdentityKeepsRetriesAndNewRequestsSeparate(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()
	req := AssignRequest{ComponentID: "task", AgentID: "agent", RuntimeProvider: "codex", Prompt: "first", RequestID: "request-1", WorkspaceID: "workspace", CreatedBy: "author"}
	first, err := store.AssignTaskAdditive(ctx, "task", req)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := store.AssignTaskAdditive(ctx, "task", req)
	if err != nil || retry.ID != first.ID {
		t.Fatalf("retry=%+v err=%v", retry, err)
	}
	req.RequestID = "request-2"
	req.Prompt = "second instruction"
	second, err := store.AssignTaskAdditive(ctx, "task", req)
	if err != nil || second.ID == first.ID || second.Prompt != req.Prompt {
		t.Fatalf("new request lost: %+v %v", second, err)
	}
	if second.State != AssignmentQueued || second.ReplacesAssignmentID != "" {
		t.Fatalf("unexpected replacement: %+v", second)
	}
	req.RequestID = "request-1"
	req.CreatedBy = "other-author"
	third, err := store.AssignTaskAdditive(ctx, "task", req)
	if err != nil || third.ID == first.ID {
		t.Fatalf("identity not scoped: %+v %v", third, err)
	}
}

func TestAdditiveRequestIdentitySurvivesReplayAndCompletion(t *testing.T) {
	req := AssignRequest{ComponentID: "task", AgentID: "agent", RuntimeProvider: "codex", Prompt: "work", RequestID: "request", WorkspaceID: "workspace", CreatedBy: "author"}
	store := NewStore()
	defer store.Close()
	state := newStoreState()
	first, err := store.handleAssign(&state, "task", req, true, false)
	if err != nil {
		t.Fatal(err)
	}
	first.State = AssignmentCompleted
	state.assignments[first.ID] = first
	restored, err := stateFromSnapshot(snapshotFromState(&state, store.now()))
	if err != nil {
		t.Fatal(err)
	}
	retry, err := store.handleAssign(&restored, "task", req, true, false)
	if err != nil || retry.ID != first.ID || retry.State != AssignmentCompleted {
		t.Fatalf("replayed request=%+v %v", retry, err)
	}
}

func TestRequestMembershipPersistsAfterCompletionAndCancelsEveryQueuedRequest(t *testing.T) {
	ctx := context.Background()
	client := NewDevelopmentAIAgentClientStore()
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: client})
	defer store.Close()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev-riid"}
	server := NewServer(ServerConfig{AIAgentClient: client, Assignment: store})
	var ids []string
	for _, requestID := range []string{"one", "two"} {
		assignment, err := store.AssignTaskAdditive(ctx, "task-many", AssignRequest{ComponentID: "task-many", AgentID: "agent-public-openclaw", RuntimeProvider: "openclaw", Prompt: "work", RequestID: requestID, WorkspaceID: principal.WorkspaceID, CreatedBy: principal.PrincipalID})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, assignment.ID)
		_, err = client.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-many", AssignAIAgentTaskRequest{AgentID: assignment.AgentID, RequestID: requestID, AssignmentID: assignment.ID})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := server.cancelAIAgentAssignmentBeforeAction(ctx, principal, "task-many", "agent-public-openclaw", "", "removed"); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		projection, found, err := store.LoadAssignmentProjection(ctx, id)
		if err != nil || !found || projection.Assignment.State != AssignmentCancelled {
			t.Fatalf("request still executable: %+v %v", projection, err)
		}
	}
	threads, err := client.ListAIAgentTaskThreads(ctx, principal, "task-many")
	if err != nil || len(threads.AssignedAgentIDs) != 1 {
		t.Fatalf("membership lost: %+v %v", threads, err)
	}
}

func TestReplayedRequestDoesNotRestoreExplicitlyRemovedMembership(t *testing.T) {
	ctx := context.Background()
	client := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev-riid"}
	req := AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw", AssignmentID: "assignment", RequestID: "request"}
	first, err := client.CreateAIAgentTaskAgentAssignment(ctx, principal, "task", req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.UnassignAIAgentTask(ctx, principal, "task", UnassignAIAgentTaskRequest{AgentID: req.AgentID}); err != nil {
		t.Fatal(err)
	}
	replay, err := client.CreateAIAgentTaskAgentAssignment(ctx, principal, "task", req)
	if err != nil || replay.AssignmentID != first.AssignmentID {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	threads, err := client.ListAIAgentTaskThreads(ctx, principal, "task")
	if err != nil || len(threads.AssignedAgentIDs) != 0 || len(threads.Threads) != 1 {
		t.Fatalf("restored removed membership: %+v %v", threads, err)
	}
}
