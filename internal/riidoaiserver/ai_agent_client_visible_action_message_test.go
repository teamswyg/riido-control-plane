package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentActionResponsesHideQueuedButUseTerminalMessages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	assigned, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-visible", AssignAIAgentTaskRequest{
		AgentID: "agent-public-openclaw",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment: %v", err)
	}

	queued, err := store.SubmitAIAgentTaskComment(ctx, principal, "task-visible", SubmitAIAgentTaskCommentRequest{
		AgentID: assigned.AgentID,
		Body:    "continue",
	})
	if err != nil {
		t.Fatalf("SubmitAIAgentTaskComment: %v", err)
	}
	if queued.Message != "" ||
		queued.ResultMessage != "" ||
		queued.CommentKind != "" ||
		queued.AssignmentState != AgentAssignmentStateQueued ||
		queued.WorkStatus != AgentWorkStatusIdle {
		t.Fatalf("queued response should hide copy but preserve lifecycle: %+v", queued)
	}

	stopped, err := store.StopAIAgentTask(ctx, principal, "task-visible", StopAIAgentTaskRequest{
		AgentID: queued.AgentID,
	})
	if err != nil {
		t.Fatalf("StopAIAgentTask: %v", err)
	}
	if stopped.Message != clientMessageTaskStopped {
		t.Fatalf("stopped message = %q", stopped.Message)
	}
}
