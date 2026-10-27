package riidoaiserver

import (
	"context"
	"testing"
)

func assertExecutionClaimRestartHistory(t *testing.T, cfg StoreConfig, records []AssignmentOperationRecord) {
	t.Helper()
	var state storeState
	if cfg.SnapshotStore != nil {
		snapshot, ok, err := cfg.SnapshotStore.LoadStoreSnapshot(context.Background())
		if err != nil || !ok {
			t.Fatal("missing snapshot")
		}
		state, err = stateFromSnapshot(snapshot)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		var err error
		state, err = stateFromAssignmentOperations(records)
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(state.executionClaims) != 1 {
		t.Fatal("terminal discarded immutable claim")
	}
}
