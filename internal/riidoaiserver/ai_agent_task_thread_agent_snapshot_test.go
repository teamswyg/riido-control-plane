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
	if _, err := store.RecordAIAgentThreadProgress(ctx, assigned.AgentID, AgentThreadProgressBatchRequest{
		AssignmentID: assigned.AssignmentID,
		TaskID:       assigned.TaskID,
		ThreadID:     assigned.ThreadID,
		RunID:        assigned.RunID,
		Lines: []AgentThreadProgressLine{{
			Seq:     1,
			Message: "삭제 전 작업 로그",
		}},
	}); err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
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
	if len(threads.Threads[0].Lines) != 1 || threads.Threads[0].Lines[0].Message != "삭제 전 작업 로그" {
		t.Fatalf("thread progress after delete = %+v", threads.Threads[0].Lines)
	}
	if threads.Threads[0].AssignmentState != AgentAssignmentStateStopped ||
		threads.Threads[0].CommentKind != AgentTaskCommentStoppedByAgentDeleted ||
		threads.Threads[0].ResultMessage == "" {
		t.Fatalf("thread terminal state after delete = %+v", threads.Threads[0])
	}
	status, ok := lastWorkStatusChangedEventForTask(t, store, principal, assigned.TaskID)
	if !ok || status.ThreadID != assigned.ThreadID || status.AssignmentID != assigned.AssignmentID ||
		status.RunID != assigned.RunID || status.CommentKind != AgentTaskCommentStoppedByAgentDeleted {
		t.Fatalf("delete event should target actual thread, got=%+v found=%v assigned=%+v", status, ok, assigned)
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

func TestAssignedAgentProfilesFallbackToThreadSnapshotForDeletedActiveAgent(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	store.mu.Lock()
	store.taskThreads["task-legacy-active-deleted"] = []AIAgentTaskThreadRecord{{
		ThreadID:        "thread-legacy-active-deleted",
		TaskID:          "task-legacy-active-deleted",
		AssignmentID:    "asn-legacy-active-deleted",
		AgentID:         "agent-deleted-active",
		RunID:           "run-legacy-active-deleted",
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{
			AgentID:             "agent-deleted-active",
			WorkspaceID:         defaultAIAgentClientWorkspaceID,
			OwnerPrincipalID:    "user-1",
			Name:                "삭제된 홍도",
			ProfileThumbnailURL: "https://cdn.riido.io/thumbnail/ai/profile/deleted.png",
			TmpColor:            "#B87EAD",
			Visibility:          AgentVisibilityPrivate,
		},
	}}
	store.mu.Unlock()
	profiles, err := store.ListWorkspaceAssignedAgentProfiles(ctx, principal)
	if err != nil {
		t.Fatalf("ListWorkspaceAssignedAgentProfiles: %v", err)
	}
	profile := profiles.AssignedAgentProfiles["task-legacy-active-deleted"]
	if profile.AvatarURL == "" || profile.TmpColor != "#B87EAD" {
		t.Fatalf("deleted active agent profile fallback = %+v all=%+v", profile, profiles.AssignedAgentProfiles)
	}
}
