package riidoaiserver

import (
	"context"
	"testing"
)

func BenchmarkRecordAIAgentThreadProgress(b *testing.B) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-progress-bench", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	seq := 0
	for b.Loop() {
		seq++
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
			b.Fatal(err)
		}
	}
}
