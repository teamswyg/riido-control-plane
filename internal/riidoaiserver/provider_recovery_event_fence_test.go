package riidoaiserver

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestRecoveryEventRejectsMissingLeaseUnauthorizedSwitchAndInvalidState(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*AgentEventRequest)
	}{
		{"lease", func(r *AgentEventRequest) { r.Metadata["lease_token"] = "wrong" }},
		{"event-key", func(r *AgentEventRequest) { delete(r.Metadata, metadatakeys.AssignmentEventKey.String()) }},
		{"scope", func(r *AgentEventRequest) { r.RuntimeID = "" }},
		{"state", func(r *AgentEventRequest) { r.State = AssignmentRunning }},
		{"provider", func(r *AgentEventRequest) { r.Metadata[assignment.RecoveryProviderKey] = "cursor" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := Assignment{ID: "asn", TaskID: "task", AgentID: "agent", RuntimeProvider: "codex", State: AssignmentLeased, LeaseToken: "lease", UpdatedAt: time.Now()}
			state := newStoreState()
			state.assignments[a.ID] = a
			store := NewStore()
			defer store.Close()
			req := recoveryEventRequest(a)
			tc.mutate(&req)
			if _, err := store.handleEvent(&state, "agent", req); err == nil {
				t.Fatal("unfenced recovery accepted")
			}
		})
	}
}

func TestRecoveryPhaseDoesNotProjectIntoAnotherTaskRun(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	thread := eventThreadRecord("task", "new-thread", "new-assignment", AgentAssignmentStateRunning)
	store.taskThreads["task"] = []AIAgentTaskThreadRecord{thread}
	recovery := (assignment.RecoveryContext{State: "cancelled", Reason: "cancelled", Provider: "codex", Circuit: "closed", OutcomeCertainty: "not_started"}).Metadata()
	err := store.RecordAIAgentAssignmentEvent(context.Background(), "agent-owned-codex", AgentEventRequest{}, TaskEvent{TaskID: "task", AssignmentID: "old-assignment", AgentID: "agent-owned-codex", Type: EventRiidoLog, State: AssignmentCancelled, Metadata: recovery, At: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if current := store.taskThreads["task"][0]; current.AssignmentID != "new-assignment" || current.AssignmentState != AgentAssignmentStateRunning {
		t.Fatal("old recovery cancelled the current task run")
	}
}
