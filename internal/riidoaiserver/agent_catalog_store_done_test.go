package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAgentCatalogStoreReturnsWhenStoreClosesWhileWaiting(t *testing.T) {
	record := AgentCatalogRecord{
		AgentID:          "agent-a",
		OwnerPrincipalID: "user-a",
		Visibility:       AgentCatalogVisibilityPrivate,
	}
	tests := []struct {
		name string
		call func(context.Context, *Store) error
	}{
		{name: "ListAgentCatalog", call: func(ctx context.Context, store *Store) error {
			_, err := store.ListAgentCatalog(ctx)
			return err
		}},
		{name: "GetAgentCatalog", call: func(ctx context.Context, store *Store) error {
			_, _, err := store.GetAgentCatalog(ctx, "agent-a")
			return err
		}},
		{name: "SaveAgentCatalog", call: func(ctx context.Context, store *Store) error {
			_, err := store.SaveAgentCatalog(ctx, record)
			return err
		}},
		{name: "DeleteAgentCatalog", call: func(ctx context.Context, store *Store) error {
			_, err := store.DeleteAgentCatalog(ctx, "agent-a")
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectAgentCatalogStoreDone(t, tt.name, tt.call)
		})
	}
}

func expectAgentCatalogStoreDone(t *testing.T, name string, call func(context.Context, *Store) error) {
	t.Helper()
	done := make(chan struct{})
	store := &Store{commands: make(chan any, 1), done: done}
	errCh := make(chan error, 1)

	go func() {
		errCh <- call(context.Background(), store)
	}()

	deadline := time.After(time.Second)
	for len(store.commands) == 0 {
		select {
		case <-deadline:
			t.Fatalf("%s did not enqueue command", name)
		default:
			time.Sleep(time.Millisecond)
		}
	}
	close(done)

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatalf("%s succeeded after store close", name)
		}
	case <-time.After(time.Second):
		t.Fatalf("%s did not return after store close", name)
	}
}
