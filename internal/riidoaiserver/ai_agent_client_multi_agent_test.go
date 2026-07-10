package riidoaiserver

import (
	"context"
	"testing"
	"time"
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

func TestAIAgentClientStopsExplicitAssignmentWithoutStoppingSiblingThread(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	agentID := "agent-public-openclaw"
	now := time.Now().UTC()

	agent := store.agents[agentID]
	agent.WorkStatus = AgentWorkStatusRunning
	agent.AssignedTaskCount = 2
	agent.Editability = AgentEditabilityBlockedAssignedTasks
	store.agents[agentID] = agent
	store.taskThreads["task-targeted-stop"] = []AIAgentTaskThreadRecord{
		{
			ThreadID:        "thread-old",
			TaskID:          "task-targeted-stop",
			AssignmentID:    "asn-old",
			AgentID:         agentID,
			RunID:           "run-old",
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			CommentKind:     AgentTaskCommentAssignmentStarted,
			StartedAt:       now,
		},
		{
			ThreadID:        "thread-new",
			TaskID:          "task-targeted-stop",
			AssignmentID:    "asn-new",
			AgentID:         agentID,
			RunID:           "run-new",
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			CommentKind:     AgentTaskCommentAssignmentStarted,
			StartedAt:       now,
		},
	}

	stopped, err := store.StopAIAgentTaskAgentAssignment(ctx, principal, "task-targeted-stop", agentID, AgentAssignmentActionRequest{
		AssignmentID: "asn-old",
		Reason:       "targeted stop",
	})
	if err != nil {
		t.Fatalf("StopAIAgentTaskAgentAssignment: %v", err)
	}
	if stopped.AssignmentID != "asn-old" || stopped.ThreadID != "thread-old" {
		t.Fatalf("stopped wrong assignment: %+v", stopped)
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, "task-targeted-stop")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	oldThread := taskThreadByAssignment(t, threads.Threads, "asn-old")
	newThread := taskThreadByAssignment(t, threads.Threads, "asn-new")
	if taskThreadHasActiveStream(oldThread) || oldThread.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("target assignment should be stopped: %+v", oldThread)
	}
	if !taskThreadHasActiveStream(newThread) || newThread.AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("sibling assignment should remain active: %+v", newThread)
	}
}

func TestAIAgentClientStopTaskResolvesAgentFromExplicitAssignment(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	now := time.Now().UTC()
	store.taskThreads["task-assignment-target"] = []AIAgentTaskThreadRecord{
		{
			ThreadID:        "thread-openclaw",
			TaskID:          "task-assignment-target",
			AssignmentID:    "asn-openclaw",
			AgentID:         "agent-public-openclaw",
			RunID:           "run-openclaw",
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			CommentKind:     AgentTaskCommentAssignmentStarted,
			StartedAt:       now,
		},
		{
			ThreadID:        "thread-owned",
			TaskID:          "task-assignment-target",
			AssignmentID:    "asn-owned",
			AgentID:         "agent-owned-codex",
			RunID:           "run-owned",
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			CommentKind:     AgentTaskCommentAssignmentStarted,
			StartedAt:       now.Add(time.Second),
		},
	}

	stopped, err := store.StopAIAgentTask(ctx, principal, "task-assignment-target", StopAIAgentTaskRequest{
		AssignmentID: "asn-openclaw",
		Reason:       "targeted stop without agent id",
	})
	if err != nil {
		t.Fatalf("StopAIAgentTask: %v", err)
	}
	if stopped.AgentID != "agent-public-openclaw" || stopped.AssignmentID != "asn-openclaw" {
		t.Fatalf("stop should resolve target agent from assignment id: %+v", stopped)
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, "task-assignment-target")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if thread := taskThreadByAssignment(t, threads.Threads, "asn-openclaw"); taskThreadHasActiveStream(thread) {
		t.Fatalf("target assignment should be stopped: %+v", thread)
	}
	if thread := taskThreadByAssignment(t, threads.Threads, "asn-owned"); !taskThreadHasActiveStream(thread) {
		t.Fatalf("latest sibling assignment should remain active: %+v", thread)
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

func taskThreadByAssignment(t *testing.T, threads []AIAgentTaskThreadRecord, assignmentID string) AIAgentTaskThreadRecord {
	t.Helper()
	for _, thread := range threads {
		if thread.AssignmentID == assignmentID {
			return thread
		}
	}
	t.Fatalf("missing task thread for assignment %s in %+v", assignmentID, threads)
	return AIAgentTaskThreadRecord{}
}

func TestAIAgentClientThreadProjectionKeepsDurableQueueDiagnostics(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 9, 5, 0, 0, 0, time.UTC)
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-queue-diagnostics", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-current",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	reader := mapAssignmentProjectionReader{projections: map[string]AssignmentProjection{
		assigned.AssignmentID: {
			Assignment: Assignment{
				ID:                    assigned.AssignmentID,
				TaskID:                "task-queue-diagnostics",
				AgentID:               assigned.AgentID,
				RuntimeProvider:       "openclaw",
				State:                 AssignmentQueued,
				BlockedByAssignmentID: "asn-blocker",
				CreatedAt:             now.Add(-time.Minute),
				UpdatedAt:             now.Add(-time.Minute),
			},
			LastEventSeq: 7,
		},
		"asn-blocker": {
			Assignment: Assignment{
				ID:              "asn-blocker",
				TaskID:          "task-queue-diagnostics",
				AgentID:         "agent-old",
				RuntimeProvider: "codex",
				State:           AssignmentCancelling,
				CreatedAt:       now.Add(-2 * time.Minute),
				UpdatedAt:       now.Add(-30 * time.Second),
			},
			LastEventSeq: 6,
		},
	}}
	changed, err := store.ReconcileAIAgentActiveThreadProjections(ctx, principal, "task-queue-diagnostics", reader)
	if err != nil {
		t.Fatalf("ReconcileAIAgentActiveThreadProjections: %v", err)
	}
	if !changed {
		t.Fatal("expected queued projection diagnostics to update thread")
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, "task-queue-diagnostics")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %+v", threads)
	}
	if threads.Threads[0].QueueDiagnostics != nil ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateQueued ||
		threads.Threads[0].WorkStatus != AgentWorkStatusIdle {
		t.Fatalf("client projection should hide queue diagnostics: %+v", threads)
	}
	store.mu.Lock()
	durableThread := store.taskThreads["task-queue-diagnostics"][0]
	store.mu.Unlock()
	diagnostics := durableThread.QueueDiagnostics
	if diagnostics == nil ||
		diagnostics.Reason != "blocked_by_assignment" ||
		diagnostics.BlockedByAssignmentID != "asn-blocker" ||
		diagnostics.BlockerAgentID != "agent-old" ||
		diagnostics.BlockerRuntimeProvider != "codex" ||
		diagnostics.BlockerState != AssignmentCancelling ||
		!diagnostics.BlockerUpdatedAt.Equal(now.Add(-30*time.Second)) {
		t.Fatalf("queue diagnostics = %+v", diagnostics)
	}
}

type mapAssignmentProjectionReader struct {
	projections map[string]AssignmentProjection
}

func (r mapAssignmentProjectionReader) LoadAssignmentProjection(_ context.Context, assignmentID string) (AssignmentProjection, bool, error) {
	projection, ok := r.projections[assignmentID]
	return projection, ok, nil
}

func TestAIAgentClientStaleTerminalAgentCountDoesNotQueueNextAssignment(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "stale terminal count agent",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent: %v", err)
	}

	store.mu.Lock()
	agent := store.agents[created.Agent.AgentID]
	agent.WorkStatus = AgentWorkStatusCompleted
	agent.AssignedTaskCount = 1
	agent.Editability = AgentEditabilityBlockedAssignedTasks
	store.agents[agent.AgentID] = agent
	store.taskThreads["task-stale-completed"] = []AIAgentTaskThreadRecord{{
		ThreadID:        "thread-stale-completed",
		TaskID:          "task-stale-completed",
		AssignmentID:    "asn-stale-completed",
		AgentID:         agent.AgentID,
		RunID:           "run-stale-completed",
		WorkStatus:      AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted,
		CommentKind:     AgentTaskCommentTaskCompleted,
		Message:         "agent work completed",
	}}
	store.mu.Unlock()

	bootstrap, err := store.BootstrapAIAgentClient(ctx, principal, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient: %v", err)
	}
	projected := agentByID(bootstrap.Agents, agent.AgentID)
	if projected.AssignedTaskCount != 0 || projected.Editability != AgentEditabilityEditable {
		t.Fatalf("stale terminal count should be projected editable: %+v", projected)
	}

	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-after-stale", AssignAIAgentTaskRequest{AgentID: agent.AgentID})
	if err != nil {
		t.Fatalf("AssignAIAgentTask after stale terminal count: %v", err)
	}
	if assigned.WorkStatus == AgentWorkStatusQueued || assigned.AssignmentState == AgentAssignmentStateQueued {
		t.Fatalf("next assignment should start instead of queueing: %+v", assigned)
	}
}

func TestAIAgentClientDeviceDaemonResolvesThroughOwnedAgentRuntime(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	store.mu.Lock()
	for index := range store.devices {
		if store.devices[index].DeviceID != "device-dev-macbook" {
			continue
		}
		store.devices[index].OwnerPrincipalID = "device-owner-projection-lag"
		for runtimeIndex := range store.devices[index].Runtimes {
			store.devices[index].Runtimes[runtimeIndex].OwnerPrincipalID = "device-owner-projection-lag"
		}
	}
	daemon := store.daemons["device-dev-macbook"]
	daemon.OwnerPrincipalID = "device-owner-projection-lag"
	store.daemons["device-dev-macbook"] = daemon
	store.mu.Unlock()

	detail, err := store.GetAIAgentDeviceDaemon(ctx, principal, "device-dev-macbook")
	if err != nil {
		t.Fatalf("GetAIAgentDeviceDaemon through owned agent runtime: %v", err)
	}
	if detail.Daemon.DeviceID != "device-dev-macbook" || detail.Daemon.DaemonID != "daemon-dev-macbook" {
		t.Fatalf("device daemon detail = %+v", detail.Daemon)
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
