package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func mustAssignCopywritingTask(t *testing.T, store *DevelopmentAIAgentClientStore, ctx context.Context, principal AuthorizationResult) AIAgentTaskActionResponse {
	t.Helper()
	root, err := store.AssignAIAgentTask(ctx, principal, "task-copywriting", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex", AssignmentID: "asn-copy-root",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	return root
}

func recordThreadProgress(t *testing.T, store *DevelopmentAIAgentClientStore, action AIAgentTaskActionResponse, messages ...string) {
	t.Helper()
	lines := make([]AgentThreadProgressLine, 0, len(messages))
	for i, message := range messages {
		lines = append(lines, AgentThreadProgressLine{Seq: i + 1, Message: message})
	}
	res, err := store.RecordAIAgentThreadProgress(context.Background(), action.AgentID, AgentThreadProgressBatchRequest{
		AssignmentID: action.AssignmentID,
		TaskID:       action.TaskID,
		ThreadID:     action.ThreadID,
		RunID:        action.RunID,
		Lines:        lines,
	})
	if err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
	if res.AcceptedLines != len(messages) || res.Event.Lines[0].ObservedAt.IsZero() {
		t.Fatalf("progress evidence missing observed_at: %+v", res)
	}
}

func recordAssignmentFailed(t *testing.T, store *DevelopmentAIAgentClientStore, action AIAgentTaskActionResponse, message string) {
	t.Helper()
	event := TaskEvent{
		TaskID:       action.TaskID,
		AssignmentID: action.AssignmentID,
		AgentID:      action.AgentID,
		Type:         EventAssignmentFailed,
		State:        AssignmentFailed,
		Message:      message,
		At:           time.Now().UTC(),
	}
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), action.AgentID, AgentEventRequest{}, event); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent failed: %v", err)
	}
}

func recordAssignmentCompleted(t *testing.T, store *DevelopmentAIAgentClientStore, action AIAgentTaskActionResponse, message string) {
	t.Helper()
	event := TaskEvent{
		TaskID:       action.TaskID,
		AssignmentID: action.AssignmentID,
		AgentID:      action.AgentID,
		Type:         EventAssignmentCompleted,
		State:        AssignmentCompleted,
		Message:      message,
		At:           time.Now().UTC(),
	}
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), action.AgentID, AgentEventRequest{}, event); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent completed: %v", err)
	}
}
