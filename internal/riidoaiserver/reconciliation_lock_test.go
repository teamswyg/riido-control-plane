package riidoaiserver

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

type callbackProjectionReader struct{ read func() }

func (r callbackProjectionReader) LoadAssignmentProjection(ctx context.Context, id string) (AssignmentProjection, bool, error) {
	r.read()
	return AssignmentProjection{}, false, nil
}
func TestPersistentReconciliationDoesNotHoldRegistryLockAcrossAssignmentRead(t *testing.T) {
	store, err := OpenPersistentAIAgentClientStore(context.Background(), nil, NewFileAIAgentClientSnapshot(filepath.Join(t.TempDir(), "agents.json")))
	if err != nil {
		t.Fatal(err)
	}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev-riid"}
	_, err = store.CreateAIAgentTaskAgentAssignment(context.Background(), principal, "task-lock", AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw", AssignmentID: "assignment-lock"})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := store.ReconcileAIAgentActiveThreadProjections(context.Background(), principal, "task-lock", callbackProjectionReader{read: func() { store.LookupAgent("agent-public-openclaw") }})
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("assignment read could not reenter registry: lock inversion blocks heartbeat and task queries")
	}
}
