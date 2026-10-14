package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func newMembershipDynamoReplica(t *testing.T, f *membershipSharedDynamoFixture) (*PersistentAIAgentClientStore, *Store) {
	t.Helper()
	snapshot := newSnapshotDynamoDBStore(t, time.Now().UTC(), f.server.URL, nil)
	t.Cleanup(func() { _ = snapshot.Close() })
	client, err := OpenPersistentAIAgentClientStore(context.Background(), NewDevelopmentAIAgentClientStore(), snapshot)
	if err != nil {
		t.Fatalf("open replica client: %v", err)
	}
	operations := newAssignmentOperationStoreHarnessStore(t, dynamoDBAssignmentOperationStoreHarnessConfig{AccessKeyID: "AKID", TableName: "riido-ai-agent-development"}, f.server.URL)
	store, err := OpenStoreWithConfig(context.Background(), StoreConfig{AgentRegistry: client, OperationStore: operations})
	if err != nil {
		t.Fatalf("open replica assignments: %v", err)
	}
	t.Cleanup(store.Close)
	return client, store
}

// This catches process-local fencing by using two separate Persistent client
// stores, assignment actors and native Dynamo adapters with a shared table.
func TestHTTPMembershipRemovalIndependentStoresPreservesDurableUnpublishedExecution(t *testing.T) {
	for _, mode := range []string{"replacement", "additive", "reply"} {
		for _, state := range []string{"queued", "completed"} {
			for _, history := range []string{"observed", "idless", "compacted"} {
				if mode == "reply" && history == "compacted" {
					continue
				}
				t.Run(mode+"/"+state+"/"+history, func(t *testing.T) { checkMembershipIndependentStorePublication(t, mode, state, history) })
			}
		}
	}
	t.Run("initially_absent", func(t *testing.T) { checkMembershipIndependentStorePublication(t, "replacement", "queued", "absent") })
}
