package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStoreActorPersistsResumeAndProviderSessionIDs(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 16, 13, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "continue",
		ResumeSessionID: "sess-prev",
		Worktree: &AssignmentWorktree{
			RepositoryFullName: " teamswyg/riido-daemon ",
			RepositoryURL:      " https://github.com/teamswyg/riido-daemon ",
			BranchName:         " RIID-4964-agent-profile-upload ",
			Source:             " connected_pull_request ",
		},
		AgentInstruction: "resume if possible",
	})
	assertResumeAssignment(t, assignment)

	now = now.Add(time.Second)
	poll := mustPollActor(t, store, ctx, "agent-1")
	if poll.Assignment == nil || poll.Assignment.ResumeSessionID != "sess-prev" ||
		poll.Assignment.Worktree == nil || poll.Assignment.Worktree.BranchName != "RIID-4964-agent-profile-upload" {
		t.Fatalf("poll assignment did not preserve resume_session_id: %+v", poll.Assignment)
	}

	now = now.Add(time.Second)
	pinned := mustRecordActorEvent(t, store, ctx, "agent-1", AgentEventRequest{
		AssignmentID:      assignment.ID,
		DaemonID:          "daemon-1",
		RuntimeID:         "runtime-1",
		EventType:         EventProviderSessionPinned,
		ProviderSessionID: "sess-current",
		Message:           "provider session pinned",
	})
	if pinned.Assignment == nil || pinned.Assignment.ProviderSessionID != "sess-current" {
		t.Fatalf("provider_session_id was not persisted: %+v", pinned.Assignment)
	}

	now = now.Add(time.Second)
	heartbeat, err := store.HeartbeatAgent(ctx, "agent-1", AgentHeartbeatRequest{
		DaemonID: "daemon-1", RuntimeID: "runtime-1", ActiveAssignmentIDs: []string{assignment.ID},
	})
	if err != nil {
		t.Fatalf("HeartbeatAgent: %v", err)
	}
	if len(heartbeat.RefreshedAssignments) != 1 || heartbeat.RefreshedAssignments[0].ProviderSessionID != "sess-current" {
		t.Fatalf("heartbeat did not return provider_session_id: %+v", heartbeat.RefreshedAssignments)
	}
}
