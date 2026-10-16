package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLegacyReadyOperationFailureRetriesPersistenceWithoutFalseACK(t *testing.T) {
	ops := &legacyEventOperations{err: errors.New("private-operation-marker")}
	store, server, req := legacyReadyFixture(t, StoreConfig{OperationStore: ops})
	for range 2 {
		response := recordAgentEventHTTP(t, server, req)
		if response.Code != 503 || strings.Contains(response.Body.String(), "private-operation-marker") {
			t.Fatal("legacy operation failure did not retain safe retry semantics")
		}
	}
	projection, _, err := store.LoadAssignmentProjection(context.Background(), req.AssignmentID)
	if err != nil || projection.Assignment.State != AssignmentLeased || ops.calls != 2 || len(ops.saved) != 0 {
		t.Fatal("failed operation changed execution or retained a false ACK")
	}
	ops.err = nil
	first, retry := recordAgentEventHTTP(t, server, req), recordAgentEventHTTP(t, server, req)
	if first.Code != 200 || retry.Code != 200 || first.Body.String() != retry.Body.String() || ops.calls != 3 || len(ops.saved) != 1 {
		t.Fatal("legacy journal retry did not commit exactly once")
	}
}

func TestLegacyDualStorageFailsClosedBeforeEitherNonAtomicWrite(t *testing.T) {
	ops, snapshot := &legacyEventOperations{}, &legacyEventSnapshot{err: errors.New("snapshot cannot commit")}
	store, server, req := legacyReadyFixture(t, StoreConfig{OperationStore: ops, SnapshotStore: snapshot})
	for range 2 {
		if response := recordAgentEventHTTP(t, server, req); response.Code != 503 {
			t.Fatal("non-atomic legacy dual storage was accepted")
		}
	}
	projection, _, err := store.LoadAssignmentProjection(context.Background(), req.AssignmentID)
	if err != nil || projection.Assignment.State != AssignmentLeased || projection.LastEventSeq != 0 || ops.calls != 0 || snapshot.calls != 0 {
		t.Fatal("legacy dual storage partially wrote or changed state")
	}
}
