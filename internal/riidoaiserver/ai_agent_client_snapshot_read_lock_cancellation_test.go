package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCancelledAgentReadsLeaveSnapshotLockQueue(t *testing.T) {
	assignment := NewStore()
	t.Cleanup(assignment.Close)

	for _, kind := range []string{"thread projection reconciliation", "client bootstrap"} {
		t.Run(kind, func(t *testing.T) {
			store := &PersistentAIAgentClientStore{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore()}
			store.operationMu.Lock()
			defer store.operationMu.Unlock()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			result := make(chan error, 1)
			go func() {
				principal := AuthorizationResult{PrincipalID: "user-1"}
				if kind == "client bootstrap" {
					_, err := store.BootstrapAIAgentClient(ctx, principal, "")
					result <- err
					return
				}
				_, err := store.ReconcileAIAgentActiveThreadProjections(ctx, principal, "", assignment)
				result <- err
			}()

			select {
			case err := <-result:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("cancelled read returned %v", err)
				}
			case <-time.After(250 * time.Millisecond):
				t.Fatal("cancelled read waited for the snapshot operation lock")
			}
		})
	}
}
