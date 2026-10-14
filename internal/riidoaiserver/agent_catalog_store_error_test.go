package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAgentCatalogStoreReturnsErrorAfterClose(t *testing.T) {
	store := NewStore()
	store.Close()
	ctx := context.Background()
	record := AgentCatalogRecord{
		AgentID:          "agent-a",
		OwnerPrincipalID: "user-a",
		Visibility:       AgentCatalogVisibilityPrivate,
	}

	if _, err := store.ListAgentCatalog(ctx); err == nil {
		t.Fatalf("ListAgentCatalog succeeded after close")
	}
	if _, _, err := store.GetAgentCatalog(ctx, "agent-a"); err == nil {
		t.Fatalf("GetAgentCatalog succeeded after close")
	}
	if _, err := store.SaveAgentCatalog(ctx, record); err == nil {
		t.Fatalf("SaveAgentCatalog succeeded after close")
	}
	if _, err := store.DeleteAgentCatalog(ctx, "agent-a"); err == nil {
		t.Fatalf("DeleteAgentCatalog succeeded after close")
	}
}

func TestAgentCatalogStoreReturnsContextTimeoutWithoutActorReply(t *testing.T) {
	record := AgentCatalogRecord{
		AgentID:          "agent-a",
		OwnerPrincipalID: "user-a",
		Visibility:       AgentCatalogVisibilityPrivate,
	}

	expectAgentCatalogStoreTimeout(t, "ListAgentCatalog", func(ctx context.Context, store *Store) error {
		_, err := store.ListAgentCatalog(ctx)
		return err
	})
	expectAgentCatalogStoreTimeout(t, "GetAgentCatalog", func(ctx context.Context, store *Store) error {
		_, _, err := store.GetAgentCatalog(ctx, "agent-a")
		return err
	})
	expectAgentCatalogStoreTimeout(t, "SaveAgentCatalog", func(ctx context.Context, store *Store) error {
		_, err := store.SaveAgentCatalog(ctx, record)
		return err
	})
	expectAgentCatalogStoreTimeout(t, "DeleteAgentCatalog", func(ctx context.Context, store *Store) error {
		_, err := store.DeleteAgentCatalog(ctx, "agent-a")
		return err
	})
}

func expectAgentCatalogStoreTimeout(t *testing.T, name string, call func(context.Context, *Store) error) {
	t.Helper()
	store := &Store{commands: make(chan any, 1), done: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := call(ctx, store); err == nil {
		t.Fatalf("%s succeeded without actor reply", name)
	}
}
