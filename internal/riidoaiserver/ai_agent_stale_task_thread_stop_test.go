package riidoaiserver

import (
	"testing"
	"time"
)

func TestRepairStaleActiveTaskThreadsStopsOldStoppingThread(t *testing.T) {
	now := time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)
	store := staleTaskThreadRepairFixture(now, AgentAssignmentStateStopping)

	if !store.repairStaleActiveTaskThreads(now) {
		t.Fatal("repairStaleActiveTaskThreads changed=false, want true")
	}
	thread := store.taskThreads["task-stale"][0]
	if thread.AssignmentState != AgentAssignmentStateStopped ||
		thread.CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("thread after repair = %+v", thread)
	}
}
