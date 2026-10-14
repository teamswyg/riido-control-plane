package riidoaiserver

import (
	"testing"
	"time"
)

func TestRepairStaleActiveTaskThreadsClosesOldRunningThread(t *testing.T) {
	now := time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)
	store := staleTaskThreadRepairFixture(now, AgentAssignmentStateRunning)

	if !store.repairStaleActiveTaskThreads(now) {
		t.Fatal("repairStaleActiveTaskThreads changed=false, want true")
	}

	thread := store.taskThreads["task-stale"][0]
	if thread.AssignmentState != AgentAssignmentStateFailed ||
		thread.WorkStatus != AgentWorkStatusFailed ||
		thread.CommentKind != AgentTaskCommentTaskFailed ||
		thread.CompletedAt.IsZero() {
		t.Fatalf("thread after repair = %+v", thread)
	}
	agent := store.agents["agent-stale"]
	if agent.AssignedTaskCount != 0 || agent.WorkStatus != AgentWorkStatusIdle {
		t.Fatalf("agent after repair = %+v", agent)
	}
}

func TestRepairStaleActiveTaskThreadsLeavesFreshThreadActive(t *testing.T) {
	now := time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)
	store := staleTaskThreadRepairFixture(now, AgentAssignmentStateRunning)
	store.taskThreads["task-stale"][0].StartedAt = now.Add(-time.Hour)

	if store.repairStaleActiveTaskThreads(now) {
		t.Fatal("repairStaleActiveTaskThreads changed=true, want false")
	}
	if thread := store.taskThreads["task-stale"][0]; !taskThreadHasActiveStream(thread) {
		t.Fatalf("fresh thread was closed: %+v", thread)
	}
}

func staleTaskThreadRepairFixture(now time.Time, state AgentAssignmentState) *DevelopmentAIAgentClientStore {
	return &DevelopmentAIAgentClientStore{
		agents: map[string]AgentClientRecord{
			"agent-stale": {
				AgentID:           "agent-stale",
				WorkStatus:        AgentWorkStatusRunning,
				AssignedTaskCount: 1,
			},
		},
		taskThreads: map[string][]AIAgentTaskThreadRecord{
			"task-stale": {
				{
					ThreadID:        "thread-stale",
					TaskID:          "task-stale",
					AgentID:         "agent-stale",
					WorkStatus:      AgentWorkStatusRunning,
					AssignmentState: state,
					CommentKind:     AgentTaskCommentRuntimeProgress,
					StartedAt:       now.Add(-25 * time.Hour),
				},
			},
		},
	}
}
