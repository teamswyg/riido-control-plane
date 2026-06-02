package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIAgentTaskContextClientRequestsOpenAPIEndpoint(t *testing.T) {
	var gotPath string
	var gotAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get(AIAgentTaskContextHeaderWorkspaceAPIKey)
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(aiAgentTaskContextHTTPFixture())
	}))
	defer server.Close()

	client, err := NewAIAgentTaskContextClient(AIAgentTaskContextClientConfig{
		BaseURL:         server.URL,
		WorkspaceID:     "workspace-a",
		TeamID:          "RIID",
		WorkspaceAPIKey: "workspace-key",
	})
	if err != nil {
		t.Fatalf("NewAIAgentTaskContextClient: %v", err)
	}
	contextSnapshot, err := client.GetAIAgentTaskContext(context.Background(), "component-a")
	if err != nil {
		t.Fatalf("GetAIAgentTaskContext: %v", err)
	}
	if gotPath != "/workspaces/workspace-a/open-api/v1/teams/RIID/components/component-a/ai-agent-context" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAPIKey != "workspace-key" {
		t.Fatalf("api key header = %q", gotAPIKey)
	}
	if contextSnapshot.Component.ComponentType != "task" ||
		contextSnapshot.Component.BranchName != "RIID-4800-server-task-context-http-client-assignment-prompt-wiring" ||
		contextSnapshot.Document.TiptapDocumentID != "doc-a" ||
		contextSnapshot.Hierarchy.ParentTask.Title != "Parent task" ||
		contextSnapshot.Repositories[0].FullName != "teamswyg/riido-control-plane" {
		t.Fatalf("context snapshot = %+v", contextSnapshot)
	}

	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task-a", Context: contextSnapshot})
	if err != nil {
		t.Fatalf("ComposeAIAgentAssignmentPrompt: %v", err)
	}
	for _, want := range []string{
		"branch_name: RIID-4800-server-task-context-http-client-assignment-prompt-wiring",
		"full_name: teamswyg/riido-control-plane",
		"repository_url: https://github.com/teamswyg/riido-control-plane",
		"Use the context to implement the task.",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}

func TestAIAgentPrivateTaskContextClientResolvesTeamAndComponentWithBearer(t *testing.T) {
	var gotWorkspacePath string
	var gotDetailPath string
	var gotAuthorization []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = append(gotAuthorization, r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/public/components/component-a/workspace":
			gotWorkspacePath = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "component-a",
				"componentType": "task",
				"team": map[string]any{
					"id":      "team-a",
					"teamKey": "RIID",
					"workspace": map[string]any{
						"id": "workspace-a",
					},
				},
			})
		case "/teams/team-a/components/component-a":
			gotDetailPath = r.URL.Path
			if got := r.URL.Query().Get("getDocument"); got != "true" {
				t.Fatalf("getDocument = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "component-a",
				"componentType": "task",
				"title":         "Implement JWT task context",
				"keyNumber":     "RIID-4873",
				"document": map[string]any{
					"id":               "document-a",
					"tiptapDocumentId": "doc-a",
					"HTMLContent":      "<p>Use the private component document.</p>",
				},
				"milestone": map[string]any{
					"id":            "milestone-a",
					"componentType": "milestone",
					"title":         "Control plane",
					"keyNumber":     "RIID-4872",
					"project": map[string]any{
						"id":            "project-a",
						"componentType": "project",
						"title":         "AI Agent",
						"keyNumber":     "RIID-4800",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
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
	if gotWorkspacePath != "/public/components/component-a/workspace" || gotDetailPath != "/teams/team-a/components/component-a" {
		t.Fatalf("paths workspace=%q detail=%q", gotWorkspacePath, gotDetailPath)
	}
	for _, got := range gotAuthorization {
		if got != "Bearer user-jwt" {
			t.Fatalf("authorization header = %q", got)
		}
	}
	if contextSnapshot.Component.Title != "Implement JWT task context" ||
		contextSnapshot.Document.Content != "<p>Use the private component document.</p>" ||
		contextSnapshot.Document.ContentFormat != "html" ||
		contextSnapshot.Hierarchy.Project.Title != "AI Agent" ||
		contextSnapshot.Hierarchy.Milestone.Title != "Control plane" ||
		len(contextSnapshot.Repositories) != 0 {
		t.Fatalf("context snapshot = %+v", contextSnapshot)
	}
}

func TestAIAgentPrivateTaskContextClientRejectsWorkspaceMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "component-a",
			"componentType": "task",
			"team": map[string]any{
				"id": "team-a",
				"workspace": map[string]any{
					"id": "workspace-other",
				},
			},
		})
	}))
	defer server.Close()
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("NewAIAgentPrivateTaskContextClient: %v", err)
	}
	_, err = client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{
		ComponentID: "component-a",
		WorkspaceID: "workspace-a",
		BearerToken: "user-jwt",
	})
	if err == nil || !strings.Contains(err.Error(), "workspace mismatch") {
		t.Fatalf("GetAIAgentTaskContextForRequest err=%v", err)
	}
}

func TestAIAgentTaskContextClientFailsClosedOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()
	client, err := NewAIAgentTaskContextClient(AIAgentTaskContextClientConfig{
		BaseURL:         server.URL,
		WorkspaceID:     "workspace-a",
		TeamID:          "RIID",
		WorkspaceAPIKey: "workspace-key",
	})
	if err != nil {
		t.Fatalf("NewAIAgentTaskContextClient: %v", err)
	}
	_, err = client.GetAIAgentTaskContext(context.Background(), "component-a")
	if err == nil || !strings.Contains(err.Error(), "HTTP 502") {
		t.Fatalf("GetAIAgentTaskContext err=%v", err)
	}
}

func TestAIAgentTaskContextClientRejectsUnsafeConfig(t *testing.T) {
	tests := []AIAgentTaskContextClientConfig{
		{BaseURL: "", WorkspaceID: "workspace-a", TeamID: "RIID", WorkspaceAPIKey: "key"},
		{BaseURL: "ftp://api.riido.io", WorkspaceID: "workspace-a", TeamID: "RIID", WorkspaceAPIKey: "key"},
		{BaseURL: "https://api.riido.io?debug=true", WorkspaceID: "workspace-a", TeamID: "RIID", WorkspaceAPIKey: "key"},
		{BaseURL: "https://user@api.riido.io", WorkspaceID: "workspace-a", TeamID: "RIID", WorkspaceAPIKey: "key"},
		{BaseURL: "https://api.riido.io", WorkspaceID: "", TeamID: "RIID", WorkspaceAPIKey: "key"},
		{BaseURL: "https://api.riido.io", WorkspaceID: "workspace-a", TeamID: "", WorkspaceAPIKey: "key"},
		{BaseURL: "https://api.riido.io", WorkspaceID: "workspace-a", TeamID: "RIID", WorkspaceAPIKey: ""},
	}
	for _, tt := range tests {
		if _, err := NewAIAgentTaskContextClient(tt); err == nil {
			t.Fatalf("NewAIAgentTaskContextClient(%+v) expected error", tt)
		}
	}
}

func TestAIAgentTaskContextClientRejectsUnknownFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component":{"id":"component-a","title":"Task","unknown":true},"document":{"content":"body"},"hierarchy":{},"repositories":[]}`))
	}))
	defer server.Close()
	client, err := NewAIAgentTaskContextClient(AIAgentTaskContextClientConfig{
		BaseURL:         server.URL,
		WorkspaceID:     "workspace-a",
		TeamID:          "RIID",
		WorkspaceAPIKey: "workspace-key",
	})
	if err != nil {
		t.Fatalf("NewAIAgentTaskContextClient: %v", err)
	}
	_, err = client.GetAIAgentTaskContext(context.Background(), "component-a")
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("GetAIAgentTaskContext err=%v", err)
	}
}

func aiAgentTaskContextHTTPFixture() AIAgentTaskContext {
	return AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{
			ID:            "component-a",
			ComponentType: "task",
			Title:         "Implement task context wiring",
			KeyNumber:     "4800",
			BranchName:    "RIID-4800-server-task-context-http-client-assignment-prompt-wiring",
		},
		Document: AIAgentTaskContextDocument{
			ID:               "document-a",
			TiptapDocumentID: "doc-a",
			Content:          "Use the context to implement the task.",
			ContentFormat:    "markdown",
		},
		Hierarchy: AIAgentTaskContextHierarchy{
			Project:    AIAgentTaskContextReference{ID: "project-a", Title: "[v1.22] AI Contributors", KeyNumber: "4539"},
			Milestone:  AIAgentTaskContextReference{ID: "milestone-a", Title: "Riido AI Agent Policy", KeyNumber: "4719"},
			ParentTask: AIAgentTaskContextReference{ID: "parent-a", Title: "Parent task", KeyNumber: "4799"},
		},
		Repositories: []AIAgentTaskContextRepository{{
			ID:            "repo-a",
			FullName:      "teamswyg/riido-control-plane",
			IsPrivate:     false,
			RepositoryURL: "https://github.com/teamswyg/riido-control-plane",
			Source:        TaskContextRepositorySourceConnectedPullRequest,
		}},
	}
}
