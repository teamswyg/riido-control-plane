package riidoaiserver

import (
	"reflect"
	"testing"
)

func TestCompletedResultRepairPreservesClosedRunAndOtherAgentWork(t *testing.T) {
	f, first, _, _ := terminalResultFixture(t)
	seedMissingTerminalResult(t, f, first)
	before := copyTaskThread(f.client.taskThreads[first.TaskID][0])
	agent := f.client.agents[first.AgentID]
	agent.WorkStatus, agent.AssignedTaskCount = AgentWorkStatusRunning, 2
	f.client.agents[first.AgentID] = agent
	changed, err := f.client.ReconcileAIAgentActiveThreadProjections(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, first.TaskID, f.store)
	if err != nil || !changed {
		t.Fatal("repair failed", err)
	}
	after := f.client.taskThreads[first.TaskID][0]
	if !after.CompletedAt.Equal(before.CompletedAt) || !after.StartedAt.Equal(before.StartedAt) ||
		after.AssignmentID != before.AssignmentID || after.RunID != before.RunID ||
		!reflect.DeepEqual(f.client.agents[first.AgentID], agent) {
		t.Fatal("result repair changed execution identity, timestamps or another active task")
	}
}

func TestCompletedResultRepairRejectsUnrelatedOrConflictingEvidence(t *testing.T) {
	for _, mutation := range []string{"assignment", "task", "agent", "sequence", "event_type", "state", "cancelled_thread", "existing_result"} {
		t.Run(mutation, func(t *testing.T) {
			f, first, _, _ := terminalResultFixture(t)
			seedMissingTerminalResult(t, f, first)
			projection, _, err := f.store.LoadAssignmentProjection(t.Context(), first.AssignmentID)
			if err != nil || projection.TerminalEvent == nil {
				t.Fatal("terminal evidence missing", err)
			}
			switch mutation {
			case "assignment":
				projection.TerminalEvent.AssignmentID = "another-assignment"
			case "task":
				projection.TerminalEvent.TaskID = "another-task"
			case "agent":
				projection.TerminalEvent.AgentID = "another-agent"
			case "sequence":
				projection.TerminalEvent.Seq++
			case "event_type":
				projection.TerminalEvent.Type = EventProviderLog
			case "state":
				projection.TerminalEvent.State = AssignmentFailed
			case "cancelled_thread":
				f.client.taskThreads[first.TaskID][0].AssignmentState = AgentAssignmentStateStopped
			case "existing_result":
				f.client.taskThreads[first.TaskID][0].ResultMessage = "Preserve the existing answer."
			}
			before := copyTaskThread(f.client.taskThreads[first.TaskID][0])
			reader := mapAssignmentProjectionReader{projections: map[string]AssignmentProjection{first.AssignmentID: projection}}
			changed, err := f.client.ReconcileAIAgentActiveThreadProjections(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, first.TaskID, reader)
			if err != nil || changed || !reflect.DeepEqual(before, f.client.taskThreads[first.TaskID][0]) {
				t.Fatalf("unrelated evidence mutated thread: changed=%v error=%v", changed, err)
			}
		})
	}
}
