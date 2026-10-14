package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAIAgentClientTerminalStatusFansOutProgressCompatibleEvent(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-terminal-live", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-terminal-live",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	_, live, cancel, err := store.SubscribeAIAgentClientEvents(ctx, principal)
	if err != nil {
		t.Fatalf("SubscribeAIAgentClientEvents: %v", err)
	}
	defer cancel()

	err = store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, TaskEvent{
		TaskID:       assigned.TaskID,
		AssignmentID: assigned.AssignmentID,
		AgentID:      assigned.AgentID,
		Type:         EventAssignmentCompleted,
		State:        AssignmentCompleted,
		Message:      "done",
		At:           time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent: %v", err)
	}
	event := awaitTerminalProgressEvent(t, live)
	if event.AssignmentID != assigned.AssignmentID ||
		event.ThreadID != assigned.ThreadID ||
		event.RunID != assigned.RunID ||
		event.WorkStatus != AgentWorkStatusCompleted ||
		event.AssignmentState != AgentAssignmentStateCompleted ||
		event.CommentKind != AgentTaskCommentTaskCompleted ||
		len(event.Lines) != 0 {
		t.Fatalf("terminal progress event = %+v", event)
	}
}

func awaitTerminalProgressEvent(t *testing.T, live <-chan ClientStreamEvent) AgentThreadProgressEvent {
	t.Helper()
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		select {
		case event := <-live:
			progress, ok := event.Payload.(AgentThreadProgressEvent)
			if ok && progress.AssignmentState == AgentAssignmentStateCompleted {
				return progress
			}
		case <-timer.C:
			t.Fatal("timed out waiting for terminal progress-compatible event")
		}
	}
}
