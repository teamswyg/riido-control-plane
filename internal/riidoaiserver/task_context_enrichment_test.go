package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTaskContextEnrichmentUsesAuthorizedParentsAndResourceMetadataOnly(t *testing.T) {
	var mu sync.Mutex
	reads := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fixture-jwt" {
			t.Error("request bearer lost")
			http.Error(w, "denied", http.StatusForbidden)
			return
		}
		mu.Lock()
		reads[r.URL.Path]++
		mu.Unlock()
		if r.URL.Path == "/public/components/task/workspace" {
			writeTaskContextFixture(t, w, map[string]any{"id": "task", "team": map[string]any{"id": "team", "workspace": map[string]any{"id": "workspace"}}})
			return
		}
		if r.URL.Path == "/teams/team/components/task" {
			writeTaskContextFixture(t, w, map[string]any{"id": "task", "componentType": "subtask", "title": "Fix regression", "task": map[string]any{"id": "parent", "title": "Parent"}, "document": map[string]any{"HTMLContent": "<p>Task criteria</p>"}})
			return
		}
		if r.URL.Path == "/teams/team/components/parent" {
			writeTaskContextFixture(t, w, map[string]any{"id": "parent", "componentType": "task", "title": "Parent", "project": map[string]any{"id": "project", "title": "Product"}, "milestone": map[string]any{"id": "milestone", "title": "Goal"}, "document": map[string]any{"HTMLContent": "<p>Parent criteria</p>"}})
			return
		}
		if r.URL.Path == "/teams/team/components/project" {
			writeTaskContextFixture(t, w, map[string]any{"id": "project", "document": map[string]any{"HTMLContent": "<p>Product constraints</p>"}})
			return
		}
		if r.URL.Path == "/teams/team/components/milestone" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if r.URL.Path == "/teams/team/components/task/files" {
			writeTaskContextFixture(t, w, []map[string]any{{"fileName": "spec.pdf", "downloadUrl": "https://cdn.example.test/spec.pdf"}})
			return
		}
		if r.URL.Path == "/teams/team/components/task/web-links" {
			writeTaskContextFixture(t, w, map[string]any{"links": []map[string]any{{"linkName": "Design", "linkUrl": "https://docs.example.test/design"}}})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	got, err := client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace", BearerToken: "fixture-jwt"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Hierarchy.Project.Document == nil || !strings.Contains(got.Hierarchy.Project.Document.Content, "Product constraints") || got.Hierarchy.Milestone.Document != nil || got.Hierarchy.Milestone.DocumentStatus != "unavailable" {
		t.Fatalf("parent availability = %+v", got.Hierarchy)
	}
	if len(got.Resources) != 2 || got.ResourceStatus != "available" {
		t.Fatalf("resources = %+v, status=%s", got.Resources, got.ResourceStatus)
	}
	if reads["/teams/team/components/parent"] != 1 {
		t.Fatal("parent hierarchy retrieval was not reused")
	}
	for path := range reads {
		if strings.Contains(path, "spec.pdf") || strings.Contains(path, "design") {
			t.Fatal("linked contents fetched")
		}
	}
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task", Context: got})
	if err != nil || !strings.Contains(prompt.Prompt, "spec.pdf") || !strings.Contains(prompt.Prompt, "contents have not been downloaded") {
		t.Fatal("resource evidence missing")
	}
}

func TestParentContextRejectsMismatchedIdentityAndHonorsDeadline(t *testing.T) {
	clientServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "wrong") {
			writeTaskContextFixture(t, w, map[string]any{"id": "foreign", "document": map[string]any{"HTMLContent": "foreign-private-data"}})
			return
		}
		<-r.Context().Done()
	}))
	defer clientServer.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: clientServer.URL})
	snapshot := AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task"}, Hierarchy: AIAgentTaskContextHierarchy{Project: AIAgentTaskContextReference{ID: "wrong"}, Milestone: AIAgentTaskContextReference{ID: "slow"}}}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	started := time.Now()
	client.enrichParentTaskContext(ctx, "team", "task", "fixture-jwt", &snapshot)
	if time.Since(started) > 300*time.Millisecond {
		t.Fatal("parent read exceeded request deadline")
	}
	if snapshot.Hierarchy.Project.Document != nil || snapshot.Hierarchy.Milestone.Document != nil {
		t.Fatal("unavailable parent content leaked")
	}
}

func writeTaskContextFixture(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("write fixture: %v", err)
	}
}
