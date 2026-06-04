package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentClientAdditiveAssignmentsExposeActiveThreadFilters(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "multi agent frontend",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent: %v", err)
	}

	first, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-multi", AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw"})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment first: %v", err)
	}
	second, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-multi", AssignAIAgentTaskRequest{AgentID: created.Agent.AgentID})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment second: %v", err)
	}
	if first.ThreadID == second.ThreadID {
		t.Fatalf("multi-agent assignments must create distinct task threads: first=%+v second=%+v", first, second)
	}

	threads, err := store.ListAIAgentTaskThreads(ctx, principal, "task-multi")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 2 {
		t.Fatalf("threads = %+v", threads)
	}
	if threads.ActiveStream == nil || threads.ActiveStream.ThreadID != second.ThreadID {
		t.Fatalf("legacy active_stream should stay singular and target latest active thread: %+v", threads.ActiveStream)
	}

	subscription, err := store.GetAIAgentTaskThreadStreamSubscription(ctx, principal, "task-multi")
	if err != nil {
		t.Fatalf("GetAIAgentTaskThreadStreamSubscription: %v", err)
	}
	if subscription.Stream.Href != "/v2/client/workspaces/"+defaultAIAgentClientWorkspaceID+"/ai-agent/events" {
		t.Fatalf("subscription stream = %+v", subscription.Stream)
	}
	if len(subscription.ActiveThreadFilters) != 2 {
		t.Fatalf("active_thread_filters = %+v", subscription.ActiveThreadFilters)
	}
	if !hasThreadFilter(subscription.ActiveThreadFilters, first.AgentID, first.ThreadID, first.RunID) {
		t.Fatalf("missing first active filter: %+v", subscription.ActiveThreadFilters)
	}
	if !hasThreadFilter(subscription.ActiveThreadFilters, second.AgentID, second.ThreadID, second.RunID) {
		t.Fatalf("missing second active filter: %+v", subscription.ActiveThreadFilters)
	}

	stopped, err := store.StopAIAgentTaskAgentAssignment(ctx, principal, "task-multi", first.AgentID, AgentAssignmentActionRequest{Reason: "test stop one agent"})
	if err != nil {
		t.Fatalf("StopAIAgentTaskAgentAssignment: %v", err)
	}
	if stopped.AgentID != first.AgentID || stopped.ThreadID != first.ThreadID {
		t.Fatalf("stopped wrong assignment: %+v", stopped)
	}
	bootstrap, err := store.BootstrapAIAgentClient(ctx, principal, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient after stop: %v", err)
	}
	if agent := agentByID(bootstrap.Agents, first.AgentID); agent.AssignedTaskCount != 0 || agent.Editability != AgentEditabilityEditable {
		t.Fatalf("stopped agent should be editable again: %+v", agent)
	}
	subscription, err = store.GetAIAgentTaskThreadStreamSubscription(ctx, principal, "task-multi")
	if err != nil {
		t.Fatalf("GetAIAgentTaskThreadStreamSubscription after stop: %v", err)
	}
	if len(subscription.ActiveThreadFilters) != 1 || !hasThreadFilter(subscription.ActiveThreadFilters, second.AgentID, second.ThreadID, second.RunID) {
		t.Fatalf("active_thread_filters after one stop = %+v", subscription.ActiveThreadFilters)
	}
}

func TestAIAgentClientLegacyAssignStillStopsExistingTaskThreads(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	first, err := store.AssignAIAgentTask(ctx, principal, "task-legacy-single", AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw"})
	if err != nil {
		t.Fatalf("AssignAIAgentTask first: %v", err)
	}
	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "legacy replacement frontend",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent: %v", err)
	}
	second, err := store.AssignAIAgentTask(ctx, principal, "task-legacy-single", AssignAIAgentTaskRequest{AgentID: created.Agent.AgentID})
	if err != nil {
		t.Fatalf("AssignAIAgentTask second: %v", err)
	}

	threads, err := store.ListAIAgentTaskThreads(ctx, principal, "task-legacy-single")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 2 {
		t.Fatalf("threads = %+v", threads)
	}
	if threads.ActiveStream == nil || threads.ActiveStream.ThreadID != second.ThreadID {
		t.Fatalf("active stream should target replacement thread: %+v", threads.ActiveStream)
	}
	for _, thread := range threads.Threads {
		if thread.ThreadID == first.ThreadID && taskThreadHasActiveStream(thread) {
			t.Fatalf("legacy first thread should have been stopped: %+v", thread)
		}
	}
}

func hasThreadFilter(filters []AIAgentTaskThreadStreamTarget, agentID, threadID, runID string) bool {
	for _, filter := range filters {
		if filter.AgentID == agentID && filter.ThreadID == threadID && filter.RunID == runID {
			return true
		}
	}
	return false
}

func agentByID(agents []AgentClientRecord, agentID string) AgentClientRecord {
	for _, agent := range agents {
		if agent.AgentID == agentID {
			return agent
		}
	}
	return AgentClientRecord{}
}
