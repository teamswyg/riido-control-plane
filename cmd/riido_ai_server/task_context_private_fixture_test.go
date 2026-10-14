package main

import (
	"encoding/json"
	"net/http"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func privateTaskContextRequest() riidoaiserver.AIAgentTaskContextRequest {
	return riidoaiserver.AIAgentTaskContextRequest{
		ComponentID: "component-a",
		WorkspaceID: "workspace-a",
		BearerToken: "user-jwt",
	}
}

func writePrivateTaskContextResponse(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/public/components/component-a/workspace":
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "component-a",
			"componentType": "task",
			"team":          map[string]any{"id": "team-a", "workspace": map[string]any{"id": "workspace-a"}},
		})
	case "/teams/team-a/components/component-a":
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "component-a",
			"componentType": "task",
			"title":         "Private task context from existing API server",
			"keyNumber":     "RIID-4873",
			"document": map[string]any{
				"id":               "document-a",
				"tiptapDocumentId": "doc-a",
				"HTMLContent":      "<p>Existing API server private document.</p>",
			},
		})
	case "/documents/providers/team-a/component-a":
		_ = json.NewEncoder(w).Encode(map[string]any{
			"html": "<p>Provider document from Mongo.</p>",
		})
	default:
		http.NotFound(w, r)
	}
}
