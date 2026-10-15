package riidoaiserver

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestTaskAgentParticipantRemovalPreservesTerminalHistory(t *testing.T) {
	for _, state := range []AgentAssignmentState{AgentAssignmentStateCompleted, AgentAssignmentStateFailed, AgentAssignmentStateStopped} {
		for _, scoped := range []bool{false, true} {
			t.Run(string(state)+"/"+map[bool]string{false: "agent", true: "assignment"}[scoped], func(t *testing.T) {
				ctx := context.Background()
				principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
				snapshots := &memoryAIAgentClientSnapshotStore{}
				store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
				if err != nil {
					t.Fatal(err)
				}
				assigned, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-terminal-removal", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "asn-terminal-removal"})
				if err != nil {
					t.Fatal(err)
				}
				store.mu.Lock()
				thread := &store.taskThreads[assigned.TaskID][0]
				thread.AssignmentState = state
				thread.WorkStatus = AgentWorkStatusCompleted
				thread.ResultMessage = "verified result"
				thread.Message = "original terminal message"
				thread.CompletedAt = time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
				thread.FailureDiagnostics = &AIAgentTaskThreadFailureDiagnostics{FailureCategory: "provider", Message: "original failure"}
				store.mu.Unlock()
				if _, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, assigned.TaskID, AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw"}); err != nil {
					t.Fatal(err)
				}
				before, err := store.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
				if err != nil {
					t.Fatal(err)
				}
				request := AgentAssignmentActionRequest{}
				if scoped {
					request.AssignmentID = assigned.AssignmentID
				}
				removed, err := store.DeleteAIAgentTaskAgentAssignment(ctx, principal, assigned.TaskID, assigned.AgentID, request)
				if err != nil {
					t.Fatal(err)
				}
				if removed.AssignmentState != state || removed.ResultMessage != "verified result" || removed.FailureDiagnostics == nil {
					t.Fatalf("removal response overwrote terminal data: %+v", removed)
				}
				reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
				if err != nil {
					t.Fatal(err)
				}
				after, err := reopened.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(before.Threads, after.Threads) {
					t.Fatalf("removal changed execution history: before=%+v after=%+v", before.Threads, after.Threads)
				}
				if after.ActiveStream == nil {
					t.Fatal("removal stopped peer's active stream")
				}
				ids := membershipAgentIDs(t, reopened, principal, assigned.TaskID)
				if len(ids) != 1 || ids[0] != "agent-public-openclaw" {
					t.Fatalf("removed member survived restart: %v", ids)
				}
			})
		}
	}
}
