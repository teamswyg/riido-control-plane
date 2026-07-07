package riidoaiserver

import "testing"

func TestAgentCatalogStoreHandlersValidateInputs(t *testing.T) {
	state := newStoreState()

	if _, ok, err := handleGetAgentCatalog(&state, " "); err == nil || ok {
		t.Fatalf("handleGetAgentCatalog blank id ok=%v err=%v", ok, err)
	}
	if _, err := handleSaveAgentCatalog(&state, AgentCatalogRecord{
		AgentID:    "agent-a",
		Visibility: AgentCatalogVisibilityPrivate,
	}); err == nil {
		t.Fatalf("handleSaveAgentCatalog accepted missing owner")
	}
	if deleted, err := handleDeleteAgentCatalog(&state, " "); err == nil || deleted {
		t.Fatalf("handleDeleteAgentCatalog blank id deleted=%v err=%v", deleted, err)
	}
	if deleted, err := handleDeleteAgentCatalog(&state, "missing"); err != nil || deleted {
		t.Fatalf("handleDeleteAgentCatalog missing = %v %v", deleted, err)
	}
}

func TestAgentCatalogStoreHandlersNormalizeAndSort(t *testing.T) {
	state := newStoreState()
	for _, record := range []AgentCatalogRecord{
		{AgentID: " z-agent ", OwnerPrincipalID: " user-z ", Visibility: "private "},
		{AgentID: " a-agent ", OwnerPrincipalID: " user-a ", Visibility: "public "},
	} {
		if _, err := handleSaveAgentCatalog(&state, record); err != nil {
			t.Fatalf("handleSaveAgentCatalog: %v", err)
		}
	}

	records := handleListAgentCatalog(&state)
	if got, want := agentCatalogIDs(records), []string{"a-agent", "z-agent"}; !sameStrings(got, want) {
		t.Fatalf("records = %v, want %v", got, want)
	}
	got, ok, err := handleGetAgentCatalog(&state, " z-agent ")
	if err != nil || !ok || got.OwnerPrincipalID != "user-z" {
		t.Fatalf("handleGetAgentCatalog = %+v ok=%v err=%v", got, ok, err)
	}
}
