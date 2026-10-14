package riidoaiserver

import (
	"context"
	"testing"
)

func TestAgentCatalogStoreReturnsContextErrorBeforeSend(t *testing.T) {
	record := AgentCatalogRecord{
		AgentID:          "agent-a",
		OwnerPrincipalID: "user-a",
		Visibility:       AgentCatalogVisibilityPrivate,
	}

	expectAgentCatalogStoreSendError(t, "ListAgentCatalog", func(ctx context.Context, store *Store) error {
		_, err := store.ListAgentCatalog(ctx)
		return err
	})
	expectAgentCatalogStoreSendError(t, "GetAgentCatalog", func(ctx context.Context, store *Store) error {
		_, _, err := store.GetAgentCatalog(ctx, "agent-a")
		return err
	})
	expectAgentCatalogStoreSendError(t, "SaveAgentCatalog", func(ctx context.Context, store *Store) error {
		_, err := store.SaveAgentCatalog(ctx, record)
		return err
	})
	expectAgentCatalogStoreSendError(t, "DeleteAgentCatalog", func(ctx context.Context, store *Store) error {
		_, err := store.DeleteAgentCatalog(ctx, "agent-a")
		return err
	})
}

func expectAgentCatalogStoreSendError(t *testing.T, name string, call func(context.Context, *Store) error) {
	t.Helper()
	store := &Store{commands: make(chan any), done: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := call(ctx, store); err == nil {
		t.Fatalf("%s succeeded after canceled context", name)
	}
}
