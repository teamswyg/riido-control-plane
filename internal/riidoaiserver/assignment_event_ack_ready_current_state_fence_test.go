package riidoaiserver

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestReadyACKRejectsChangedCurrentExecutionState(t *testing.T) {
	for _, state := range []AssignmentState{AssignmentRunning, AssignmentCancelling, AssignmentCancelled, AssignmentCompleted, AssignmentFailed, AssignmentQueued} {
		t.Run(string(state), func(t *testing.T) {
			ops, req := lifecycleACKFixture(t, AssignmentLeased)
			req.State, req.EventType = AssignmentReady, EventAssignmentReady
			req.Metadata[metadatakeys.AssignmentEventKey.String()] = "ready:state-fence"
			req.Metadata["lease_token"] = "lease"
			store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			defer store.Close()
			if _, err := store.RecordAgentEvent(context.Background(), "agent", req); err != nil {
				t.Fatal(err)
			}
			ops.projection.Assignment.State = state
			replica := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			defer replica.Close()
			if _, err := replica.RecordAgentEvent(context.Background(), "agent", req); err == nil || len(ops.records) != 1 {
				t.Fatalf("Ready re-authorized %s", state)
			}
		})
	}
}
