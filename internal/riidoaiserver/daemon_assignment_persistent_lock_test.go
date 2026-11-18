package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDaemonAssignmentPersistentMutexWaitHonorsContext(t *testing.T) {
	methods := []struct {
		name string
		call func(context.Context, *PersistentAIAgentClientStore) error
	}{
		{
			name: "authorize",
			call: func(ctx context.Context, store *PersistentAIAgentClientStore) error {
				var request AuthorizationRequest
				_, err := store.AuthorizeDeviceCredential(ctx, "", "", request)
				return err
			},
		},
		{
			name: "bindings",
			call: func(ctx context.Context, store *PersistentAIAgentClientStore) error {
				var principal AuthorizationResult
				_, err := store.ListAIAgentDaemonAgentBindings(ctx, principal, "")
				return err
			},
		},
	}
	for _, method := range methods {
		for _, reason := range []string{"canceled", "deadline"} {
			t.Run(method.name+"/"+reason, func(t *testing.T) {
				store := new(PersistentAIAgentClientStore)
				store.DevelopmentAIAgentClientStore = NewDevelopmentAIAgentClientStore()
				store.operationMu.Lock()
				ctx, cancel := context.WithCancel(context.Background())
				want := context.Canceled
				if reason == "deadline" {
					cancel()
					ctx, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
					want = context.DeadlineExceeded
				}
				defer cancel()
				result := make(chan error, 1)
				finished := make(chan struct{})
				defer func() {
					store.operationMu.Unlock()
					<-finished
				}()
				go func() {
					defer close(finished)
					result <- method.call(ctx, store)
				}()
				if reason == "canceled" {
					cancel()
				}
				select {
				case err := <-result:
					if !errors.Is(err, want) {
						t.Fatalf("mutex wait error = %v, want %v", err, want)
					}
				case <-time.After(time.Second):
					t.Fatal("mutex wait did not return after context ended")
				}
			})
		}
	}
}
