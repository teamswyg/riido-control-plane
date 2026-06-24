package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAIAgentTaskThreadHistoryConversationGrouping(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	first := assignConversationThread(t, store, ctx, principal, "task-conversation", "asn-first")
	markConversationThreadCompleted(store, "task-conversation", first.ThreadID, first.AgentID)
	followup, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, "task-conversation", first.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		Body:         "기존 댓글에서 이어서 해줘",
		AssignmentID: "asn-followup",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	reassigned := assignConversationThread(t, store, ctx, principal, "task-conversation", "asn-reassign")
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, "task-conversation")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	assertConversationFollowup(t, history, first.ThreadID, followup.ThreadID)
	assertConversationReassignment(t, history, first.ThreadID, reassigned.ThreadID)
}

func assignConversationThread(t *testing.T, store *DevelopmentAIAgentClientStore, ctx context.Context, principal AuthorizationResult, taskID, assignmentID string) AIAgentTaskActionResponse {
	t.Helper()
	out, err := store.AssignAIAgentTask(ctx, principal, taskID, AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex", AssignmentID: assignmentID,
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	return out
}

func markConversationThreadCompleted(store *DevelopmentAIAgentClientStore, taskID, threadID, agentID string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	threads := store.taskThreads[taskID]
	for i := range threads {
		if threads[i].ThreadID == threadID {
			threads[i].WorkStatus = AgentWorkStatusCompleted
			threads[i].AssignmentState = AgentAssignmentStateCompleted
			threads[i].CompletedAt = time.Now().UTC()
		}
	}
	agent := store.agents[agentID]
	agent.WorkStatus = AgentWorkStatusIdle
	store.agents[agentID] = agent
	store.taskThreads[taskID] = threads
}
