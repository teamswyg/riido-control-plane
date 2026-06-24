package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestStoreActorRejectsLongAgentInstruction(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	_, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:      "component-a",
		AgentID:          "agent-1",
		RuntimeProvider:  "codex",
		Prompt:           "ship it",
		AgentInstruction: strings.Repeat("지", AgentInstructionMaxCharacters+1),
	})
	if err == nil || !strings.Contains(err.Error(), "agent_instruction") {
		t.Fatalf("expected agent_instruction validation error, got %v", err)
	}
}

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
