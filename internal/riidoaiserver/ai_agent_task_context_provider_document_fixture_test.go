package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func readProviderDocumentContextFixture(t *testing.T, baseURL string) AIAgentTaskContext {
	t.Helper()
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: baseURL})
	if err != nil {
		t.Fatalf("NewAIAgentPrivateTaskContextClient: %v", err)
	}
	contextSnapshot, err := client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{
		ComponentID: "component-a",
		WorkspaceID: "workspace-a",
		BearerToken: "user-jwt",
	})
	if err != nil {
		t.Fatalf("GetAIAgentTaskContextForRequest: %v", err)
	}
	return contextSnapshot
}

func writeProviderDocumentContextFixture(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	switch r.URL.Path {
	case "/public/components/component-a/workspace":
		writeProviderDocumentWorkspace(w)
	case "/teams/team-a/components/component-a":
		writeProviderDocumentComponent(w)
	case "/documents/providers/team-a/component-a":
		if got := r.URL.Query().Get("format"); got != "html" {
			t.Fatalf("format = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"html": "<p>fresh provider document says Bye World.</p>",
		})
	default:
		http.NotFound(w, r)
	}
}

func writeProviderDocumentWorkspace(w http.ResponseWriter) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":            "component-a",
		"componentType": "task",
		"team": map[string]any{
			"id":        "team-a",
			"workspace": map[string]any{"id": "workspace-a"},
		},
	})
}

func writeProviderDocumentComponent(w http.ResponseWriter) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":            "component-a",
		"componentType": "task",
		"title":         "Implement JWT task context",
		"keyNumber":     "RIID-4873",
		"document": map[string]any{
			"id":               "document-a",
			"tiptapDocumentId": "doc-a",
			"HTMLContent":      "<p>stale component document</p>",
		},
	})
}
