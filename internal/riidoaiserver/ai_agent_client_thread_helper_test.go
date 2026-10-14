package riidoaiserver

import (
	"testing"
	"time"
)

func TestTaskThreadWorkspaceIDPrefersLiveAgent(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{agents: map[string]AgentClientRecord{
		"agent-a": {AgentID: "agent-a", WorkspaceID: " workspace-live "},
	}}
	thread := AIAgentTaskThreadRecord{
		AgentID: "agent-a",
		AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{
			WorkspaceID: "workspace-snapshot",
		},
	}

	if got := store.taskThreadWorkspaceIDLocked(thread); got != "workspace-live" {
		t.Fatalf("workspace id = %q, want live agent workspace", got)
	}
}

func TestTaskThreadWorkspaceIDFallsBackToSnapshot(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{agents: map[string]AgentClientRecord{}}
	thread := AIAgentTaskThreadRecord{
		AgentID: "agent-deleted",
		AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{
			WorkspaceID: " workspace-snapshot ",
		},
	}

	if got := store.taskThreadWorkspaceIDLocked(thread); got != "workspace-snapshot" {
		t.Fatalf("workspace id = %q, want snapshot workspace", got)
	}
}

func TestUpdateTaskThreadMessageAgentSnapshotCopiesResponse(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{workspaceID: "workspace-default"}
	thread := AIAgentTaskThreadRecord{AgentID: "agent-a"}
	now := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)

	store.updateTaskThreadMessageAgentSnapshotLocked(&thread, AIAgentTaskActionResponse{
		AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{
			AgentID: "agent-a", WorkspaceID: "workspace-a", Name: "Agent A",
		},
	}, now)

	if thread.AgentSnapshot == nil || thread.AgentSnapshot.Name != "Agent A" {
		t.Fatalf("thread snapshot = %+v", thread.AgentSnapshot)
	}
	if thread.AgentSnapshotID == "" {
		t.Fatal("snapshot id should be materialized")
	}
}

func TestUpdateTaskThreadMessageAgentSnapshotKeepsExisting(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{}
	thread := AIAgentTaskThreadRecord{AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{
		AgentID: "agent-a", Name: "Original",
	}}

	store.updateTaskThreadMessageAgentSnapshotLocked(&thread, AIAgentTaskActionResponse{
		AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{Name: "Replacement"},
	}, time.Now())

	if thread.AgentSnapshot.Name != "Original" {
		t.Fatalf("snapshot was replaced: %+v", thread.AgentSnapshot)
	}
}
