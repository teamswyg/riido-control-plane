package riidoaiserver

import (
	"context"
	"testing"
)

func TestAgentCatalogStoreActorSerializesCatalogCommands(t *testing.T) {
	store := NewStore()
	defer store.Close()
	ctx := context.Background()

	saved, err := store.SaveAgentCatalog(ctx, AgentCatalogRecord{
		AgentID:          "agent-a",
		OwnerPrincipalID: "user-a",
		Visibility:       AgentCatalogVisibilityPrivate,
	})
	if err != nil {
		t.Fatalf("SaveAgentCatalog: %v", err)
	}
	if saved.AgentID != "agent-a" || saved.Visibility != AgentCatalogVisibilityPrivate {
		t.Fatalf("saved = %+v", saved)
	}

	got, ok, err := store.GetAgentCatalog(ctx, "agent-a")
	if err != nil || !ok || got.OwnerPrincipalID != "user-a" {
		t.Fatalf("GetAgentCatalog = %+v ok=%v err=%v", got, ok, err)
	}
	records, err := store.ListAgentCatalog(ctx)
	if err != nil {
		t.Fatalf("ListAgentCatalog: %v", err)
	}
	if got, want := agentCatalogIDs(records), []string{"agent-a"}; !sameStrings(got, want) {
		t.Fatalf("records = %v, want %v", got, want)
	}

	deleted, err := store.DeleteAgentCatalog(ctx, "agent-a")
	if err != nil || !deleted {
		t.Fatalf("DeleteAgentCatalog = %v %v", deleted, err)
	}
	if _, ok, err := store.GetAgentCatalog(ctx, "agent-a"); err != nil || ok {
		t.Fatalf("deleted record still exists ok=%v err=%v", ok, err)
	}
}
