package riidoaiserver

import (
	"context"
	"testing"
)

func TestStoreActorDropsSensitiveAssignmentWorktreeURL(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
		Worktree: &AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon?token=secret",
		},
	})
	if assignment.Worktree == nil {
		t.Fatal("expected worktree")
	}
	if assignment.Worktree.RepositoryFullName != "teamswyg/riido-daemon" || assignment.Worktree.RepositoryURL != "" {
		t.Fatalf("worktree should keep full_name and drop sensitive URL: %+v", assignment.Worktree)
	}
}

func TestStoreActorDropsSensitiveAssignmentWorktreeFullName(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
		Worktree: &AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon?token=secret",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
		},
	})
	if assignment.Worktree == nil {
		t.Fatal("expected worktree")
	}
	if assignment.Worktree.RepositoryFullName != "" || assignment.Worktree.RepositoryURL != "https://github.com/teamswyg/riido-daemon" {
		t.Fatalf("worktree should drop sensitive full_name and keep safe URL: %+v", assignment.Worktree)
	}
}
