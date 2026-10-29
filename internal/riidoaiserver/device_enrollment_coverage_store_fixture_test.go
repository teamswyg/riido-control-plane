package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func openCoverageEnrollmentFixture(t *testing.T, endpoint string, enabled bool) *PersistentAIAgentClientStore {
	t.Helper()
	native := newSnapshotDynamoDBStore(t, time.Now(), endpoint, nil)
	t.Cleanup(func() { _ = native.Close() })
	base := NewDevelopmentAIAgentClientStore()
	base.ConfigureExecutionCoverageEnrollment(enabled)
	store, err := OpenPersistentAIAgentClientStore(context.Background(), base, native)
	if err != nil {
		t.Fatal(err)
	}
	return store
}
