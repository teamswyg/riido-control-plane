package riidoaiserver

import (
	"context"
	"testing"
)

func TestTaskThreadsKeepAgentSnapshotAfterAgentDeleted(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-agent-deleted", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	if assigned.AgentSnapshot == nil || assigned.AgentSnapshot.Name != "Codex 리뷰어" {
		t.Fatalf("assigned agent snapshot = %+v", assigned.AgentSnapshot)
	}
	if _, err := store.DeleteAIAgent(ctx, principal, "agent-owned-codex"); err != nil {
		t.Fatalf("DeleteAIAgent: %v", err)
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, "task-agent-deleted")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 || threads.Threads[0].ThreadID != assigned.ThreadID {
		t.Fatalf("threads after agent delete = %+v", threads)
	}
	snapshot := threads.Threads[0].AgentSnapshot
	if snapshot == nil || snapshot.AgentID != "agent-owned-codex" ||
		snapshot.ProfileThumbnailURL == "" || snapshot.RuntimeKind != RuntimeKindCodex {
		t.Fatalf("thread agent snapshot after delete = %+v", snapshot)
	}
}

func TestTaskThreadsBackfillDeletedAgentSnapshotForLegacyRows(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-legacy-agent-deleted", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	store.mu.Lock()
	threads := store.taskThreads["task-legacy-agent-deleted"]
	threads[0].AgentSnapshot = nil
	store.taskThreads["task-legacy-agent-deleted"] = threads
	delete(store.agents, "agent-owned-codex")
	store.mu.Unlock()
	threadsResponse, err := store.ListAIAgentTaskThreads(ctx, principal, "task-legacy-agent-deleted")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threadsResponse.Threads) != 1 || threadsResponse.Threads[0].ThreadID != assigned.ThreadID {
		t.Fatalf("legacy orphaned thread hidden after agent delete = %+v", threadsResponse)
	}
	snapshot := threadsResponse.Threads[0].AgentSnapshot
	if snapshot == nil || snapshot.Name != "삭제된 에이전트" || snapshot.AgentID != "agent-owned-codex" {
		t.Fatalf("legacy deleted agent snapshot = %+v", snapshot)
	}
}
