package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAIAgentTaskThreadHistoryBackfillsSnapshotIDCache(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	snapshot := &AIAgentTaskThreadAgentSnapshot{
		AgentID: "agent-1", WorkspaceID: "workspace-1", OwnerPrincipalID: "user-1",
		Name: "Agent", Visibility: AgentVisibilityPrivate, CapturedAt: time.Unix(1, 0).UTC(),
	}
	store.taskThreads["task-1"] = []AIAgentTaskThreadRecord{{
		ThreadID: "thread-1", TaskID: "task-1",
		AgentID: "agent-1", AgentSnapshot: snapshot,
		RunID: "run-1", AssignmentState: AgentAssignmentStateCompleted,
		WorkStatus: AgentWorkStatusCompleted,
	}}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-1"}
	first, err := store.ListAIAgentTaskThreadHistory(context.Background(), principal, "task-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Threads) != 1 || first.Threads[0].AgentSnapshotID == "" {
		t.Fatalf("missing response snapshot id: %+v", first.Threads)
	}
	cached := store.taskThreads["task-1"][0].AgentSnapshotID
	if cached == "" || cached != first.Threads[0].AgentSnapshotID {
		t.Fatalf("snapshot id cache = %q response=%q", cached, first.Threads[0].AgentSnapshotID)
	}
	second, err := store.ListAIAgentTaskThreadHistory(context.Background(), principal, "task-1")
	if err != nil {
		t.Fatal(err)
	}
	if second.Threads[0].AgentSnapshotID != cached || len(second.AgentSnapshots) != 1 {
		t.Fatalf("snapshot id not stable: cached=%q second=%+v", cached, second)
	}
}
