package riidoaiserver

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestAIAgentTaskThreadHistoryProgressCacheInvalidatesOnReplacement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-progress-cache", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	recordPartialProgress(t, store, assigned, 1, "thinking draft")
	assertHistoryProgressBody(t, store, principal, assigned.TaskID, "thinking draft", true)
	recordPartialProgress(t, store, assigned, 2, "thinking final")
	assertHistoryProgressBody(t, store, principal, assigned.TaskID, "thinking draft", false)
	assertHistoryProgressBody(t, store, principal, assigned.TaskID, "thinking final", true)
}

func recordPartialProgress(t *testing.T, store *DevelopmentAIAgentClientStore, assigned AIAgentTaskActionResponse, seq int, body string) {
	t.Helper()
	_, err := store.RecordAIAgentThreadProgress(context.Background(), assigned.AgentID, AgentThreadProgressBatchRequest{
		AssignmentID: assigned.AssignmentID,
		TaskID:       assigned.TaskID,
		ThreadID:     assigned.ThreadID,
		RunID:        assigned.RunID,
		Lines: []AgentThreadProgressLine{{
			Seq:        seq,
			Message:    body,
			MessageKey: progressmessage.AssistantPartialKey,
		}},
	})
	if err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
}

func assertHistoryProgressBody(
	t *testing.T,
	store *DevelopmentAIAgentClientStore,
	principal AuthorizationResult,
	taskID string,
	body string,
	want bool,
) {
	t.Helper()
	history, err := store.ListAIAgentTaskThreadHistory(context.Background(), principal, taskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	got := historyMessagesContainProgressBody(history.Threads[0].Messages, body)
	if got != want {
		t.Fatalf("progress body %q presence = %v, want %v: %+v", body, got, want, history.Threads[0].Messages)
	}
}
