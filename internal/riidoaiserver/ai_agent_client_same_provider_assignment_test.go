package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAIAgentClientSameProviderAgentsKeepDistinctAssignments(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	addSecondClaudeAgent(t, store)

	first, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-same-provider", AssignAIAgentTaskRequest{AgentID: "agent-owned-claude"})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment first: %v", err)
	}
	second, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-same-provider", AssignAIAgentTaskRequest{AgentID: "agent-owned-claude-yeongsil"})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment second: %v", err)
	}
	if first.AssignmentID == second.AssignmentID || first.ThreadID == second.ThreadID {
		t.Fatalf("same-provider assignments collapsed: first=%+v second=%+v", first, second)
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, "task-same-provider")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 2 {
		t.Fatalf("threads = %+v", threads)
	}
	taskThreadByAssignment(t, threads.Threads, first.AssignmentID)
	taskThreadByAssignment(t, threads.Threads, second.AssignmentID)

	subscription, err := store.GetAIAgentTaskThreadStreamSubscription(ctx, principal, "task-same-provider")
	if err != nil {
		t.Fatalf("GetAIAgentTaskThreadStreamSubscription: %v", err)
	}
	if !hasThreadFilter(subscription.ActiveThreadFilters, first.AgentID, first.ThreadID, first.RunID) ||
		!hasThreadFilter(subscription.ActiveThreadFilters, second.AgentID, second.ThreadID, second.RunID) {
		t.Fatalf("same-provider active_thread_filters = %+v", subscription.ActiveThreadFilters)
	}
}

func addSecondClaudeAgent(t *testing.T, store *DevelopmentAIAgentClientStore) {
	t.Helper()
	now := time.Date(2026, 6, 22, 6, 0, 0, 0, time.UTC)
	store.mu.Lock()
	defer store.mu.Unlock()
	agent := store.agents["agent-owned-claude"]
	agent.AgentID = "agent-owned-claude-yeongsil"
	agent.Name = "영실 Claude"
	agent.RuntimeID = "runtime-claude-code-yeongsil"
	agent.WorkStatus = AgentWorkStatusIdle
	agent.Editability = AgentEditabilityEditable
	agent.AssignedTaskCount = 0
	agent.CreatedAt = now
	agent.UpdatedAt = now
	store.agents[agent.AgentID] = agent
}
