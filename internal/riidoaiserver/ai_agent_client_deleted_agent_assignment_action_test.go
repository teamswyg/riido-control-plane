package riidoaiserver

import (
	"context"
	"testing"
)

func TestDeletedAgentAssignmentActionsUseThreadSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name string
		act  func(context.Context, *DevelopmentAIAgentClientStore, AuthorizationResult, AIAgentTaskActionResponse) (AIAgentTaskActionResponse, error)
	}{
		{
			name: "stop",
			act: func(ctx context.Context, store *DevelopmentAIAgentClientStore, principal AuthorizationResult, assigned AIAgentTaskActionResponse) (AIAgentTaskActionResponse, error) {
				return store.StopAIAgentTaskAgentAssignment(ctx, principal, assigned.TaskID, assigned.AgentID, AgentAssignmentActionRequest{AssignmentID: assigned.AssignmentID})
			},
		},
		{
			name: "delete_assignment",
			act: func(ctx context.Context, store *DevelopmentAIAgentClientStore, principal AuthorizationResult, assigned AIAgentTaskActionResponse) (AIAgentTaskActionResponse, error) {
				return store.DeleteAIAgentTaskAgentAssignment(ctx, principal, assigned.TaskID, assigned.AgentID, AgentAssignmentActionRequest{AssignmentID: assigned.AssignmentID})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := NewDevelopmentAIAgentClientStore()
			principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
			assigned, err := store.AssignAIAgentTask(ctx, principal, "task-deleted-agent-"+tc.name, AssignAIAgentTaskRequest{
				AgentID:      "agent-owned-codex",
				AssignmentID: "asn-deleted-agent-" + tc.name,
			})
			if err != nil {
				t.Fatalf("AssignAIAgentTask: %v", err)
			}
			if _, err := store.DeleteAIAgent(ctx, principal, assigned.AgentID); err != nil {
				t.Fatalf("DeleteAIAgent: %v", err)
			}
			response, err := tc.act(ctx, store, principal, assigned)
			if err != nil {
				t.Fatalf("deleted agent assignment action failed: %v", err)
			}
			if response.AssignmentID != assigned.AssignmentID || response.ThreadID != assigned.ThreadID {
				t.Fatalf("response = %+v, assigned = %+v", response, assigned)
			}
		})
	}
}
