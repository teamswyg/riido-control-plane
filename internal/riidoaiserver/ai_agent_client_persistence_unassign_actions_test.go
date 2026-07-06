package riidoaiserver

import (
	"context"
	"testing"
)

func TestPersistentAIAgentClientStorePersistsAssignmentRemovalActions(t *testing.T) {
	for _, tc := range []struct {
		name string
		act  func(context.Context, *PersistentAIAgentClientStore, AuthorizationResult, AIAgentTaskActionResponse) (AIAgentTaskActionResponse, error)
	}{
		{
			name: "unassign_task",
			act: func(ctx context.Context, store *PersistentAIAgentClientStore, principal AuthorizationResult, assigned AIAgentTaskActionResponse) (AIAgentTaskActionResponse, error) {
				return store.UnassignAIAgentTask(ctx, principal, assigned.TaskID, UnassignAIAgentTaskRequest{
					AgentID:      assigned.AgentID,
					AssignmentID: assigned.AssignmentID,
					Reason:       "remove participant",
				})
			},
		},
		{
			name: "delete_assignment",
			act: func(ctx context.Context, store *PersistentAIAgentClientStore, principal AuthorizationResult, assigned AIAgentTaskActionResponse) (AIAgentTaskActionResponse, error) {
				return store.DeleteAIAgentTaskAgentAssignment(ctx, principal, assigned.TaskID, assigned.AgentID, AgentAssignmentActionRequest{
					AssignmentID: assigned.AssignmentID,
					Reason:       "remove participant",
				})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			snapshots := &memoryAIAgentClientSnapshotStore{}
			principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
			store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
			if err != nil {
				t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
			}
			assigned, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-persist-"+tc.name, AssignAIAgentTaskRequest{
				AgentID:      "agent-public-openclaw",
				AssignmentID: "asn-persist-" + tc.name,
			})
			if err != nil {
				t.Fatalf("CreateAIAgentTaskAgentAssignment: %v", err)
			}
			removed, err := tc.act(ctx, store, principal, assigned)
			if err != nil {
				t.Fatalf("remove assignment: %v", err)
			}
			if removed.AssignmentState != AgentAssignmentStateStopped {
				t.Fatalf("remove response = %+v", removed)
			}

			reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
			if err != nil {
				t.Fatalf("reopen persistent store: %v", err)
			}
			threads, err := reopened.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
			if err != nil {
				t.Fatalf("ListAIAgentTaskThreads: %v", err)
			}
			if threads.ActiveStream != nil {
				t.Fatalf("removed assignment restored active stream: %+v", threads.ActiveStream)
			}
		})
	}
}
