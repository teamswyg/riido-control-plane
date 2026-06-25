package riidoaiserver

import (
	"context"
	"testing"
)

func TestRecordAIAgentThreadProgressRetainsLatestLines(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-progress-retention", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	for seq := 1; seq <= aiAgentClientThreadProgressLineLimit+25; seq++ {
		_, err := store.RecordAIAgentThreadProgress(ctx, assigned.AgentID, AgentThreadProgressBatchRequest{
			AssignmentID: assigned.AssignmentID,
			TaskID:       assigned.TaskID,
			ThreadID:     assigned.ThreadID,
			RunID:        assigned.RunID,
			Lines: []AgentThreadProgressLine{{
				Seq:     seq,
				Message: "provider fragment",
			}},
		})
		if err != nil {
			t.Fatalf("RecordAIAgentThreadProgress(%d): %v", seq, err)
		}
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	got := threads.Threads[0].Lines
	if len(got) != aiAgentClientThreadProgressLineLimit {
		t.Fatalf("line count = %d, want %d", len(got), aiAgentClientThreadProgressLineLimit)
	}
	if got[0].Seq != 26 || got[len(got)-1].Seq != aiAgentClientThreadProgressLineLimit+25 {
		t.Fatalf("retained seq range = %d..%d", got[0].Seq, got[len(got)-1].Seq)
	}
}
