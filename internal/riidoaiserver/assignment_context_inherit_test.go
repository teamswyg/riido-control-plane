package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestInheritAssignmentContextCopiesParentSnapshotForMentionedAgent(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryAssignmentContextStore()
	parent := AssignmentContextSnapshot{AssignmentID: "asg-a", TaskID: "task-1", WorkspaceID: "ws", AgentID: "agent-a", RequesterPrincipalID: "user-1", TriggerKind: AssignmentContextTriggerComment, TriggerID: "comment-r", CommentPages: 2, Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-1"}}}
	pages := [][]AIAgentTaskContextComment{{{ID: "c1"}}, {{ID: "c2"}}}
	if err := store.SaveAssignmentContextComments(ctx, parent.AssignmentID, pages); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAssignmentContextSnapshot(ctx, parent); err != nil {
		t.Fatal(err)
	}
	server := NewServer(ServerConfig{AssignmentContext: store})
	child := Assignment{ID: "asg-b", TaskID: "task-1", WorkspaceID: "ws", AgentID: "agent-b", CreatedBy: "user-1"}
	if err := server.InheritAssignmentContext(ctx, "asg-a", child, "run-a"); err != nil {
		t.Fatal(err)
	}
	got, found, err := store.LoadAssignmentContextSnapshot(ctx, "asg-b")
	if err != nil || !found || got.AgentID != "agent-b" || got.RequesterPrincipalID != "user-1" || got.TriggerKind != AssignmentContextTriggerAgentMention || got.TriggerID != "run-a" || got.CommentPages != 2 {
		t.Fatalf("child snapshot: %+v found=%v err=%v", got, found, err)
	}
	if comments, ok, _ := store.LoadAssignmentContextCommentPage(ctx, "asg-b", 1); !ok || comments[0].ID != "c2" {
		t.Fatalf("comment page not copied: %+v", comments)
	}
	if reads, _ := store.ListAssignmentContextReads(ctx, "asg-b"); len(reads) != 0 {
		t.Fatalf("inheritance must not claim prompt reads: %+v", reads)
	}
	// A second call keeps the first snapshot.
	if err := server.InheritAssignmentContext(ctx, "asg-a", child, "run-other"); err != nil {
		t.Fatal(err)
	}
	if again, _, _ := store.LoadAssignmentContextSnapshot(ctx, "asg-b"); again.TriggerID != "run-a" {
		t.Fatalf("snapshot overwritten: %+v", again)
	}
}

func TestInheritAssignmentContextRefusesOtherScopes(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryAssignmentContextStore()
	if err := store.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "asg-a", TaskID: "task-1", WorkspaceID: "ws", RequesterPrincipalID: "user-1"}); err != nil {
		t.Fatal(err)
	}
	server := NewServer(ServerConfig{AssignmentContext: store})
	for _, child := range []Assignment{
		{ID: "b1", TaskID: "task-2", WorkspaceID: "ws"},
		{ID: "b2", TaskID: "task-1", WorkspaceID: "other"},
		{ID: "b3", TaskID: "task-1", WorkspaceID: "ws", CreatedBy: "user-2"},
	} {
		if err := server.InheritAssignmentContext(ctx, "asg-a", child, "run"); !errors.Is(err, errAssignmentContextParentScope) {
			t.Fatalf("%s: err=%v", child.ID, err)
		}
	}
	if err := server.InheritAssignmentContext(ctx, "missing", Assignment{ID: "b4", TaskID: "task-1"}, "run"); !errors.Is(err, errAssignmentContextParentMissing) {
		t.Fatalf("missing parent: %v", err)
	}
	if _, err := assignmentRequesterPrincipal(Assignment{ID: "legacy"}, true); !errors.Is(err, errAgentDirectoryRequesterUnknown) {
		t.Fatalf("legacy assignment without requester must not list agents: %v", err)
	}
}
