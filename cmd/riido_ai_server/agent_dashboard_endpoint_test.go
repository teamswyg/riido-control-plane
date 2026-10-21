package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestAgentDashboardUsesSeparateGraphQLEndpoint(t *testing.T) {
	clearRiidoAIServerEnv(t)
	legacy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" {
			http.NotFound(w, r)
			return
		}
		writePrivateTaskContextResponse(w, r)
	}))
	defer legacy.Close()
	metadata := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" || r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer user-jwt" {
			t.Errorf("unexpected metadata request: %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch {
		case strings.Contains(body.Query, "AgentDashboardLocation"):
			_, _ = w.Write([]byte(`{"data":{"getComponentWorkspace":{"id":"component-a","componentType":"task","team":{"id":"team-a","workspace":{"id":"workspace-a"}}}}}`))
		case strings.Contains(body.Query, "AgentTaskProjection"):
			_, _ = w.Write([]byte(`{"data":{"structured":{"tiptapJson":"{\"type\":\"doc\",\"content\":[{\"type\":\"text\",\"text\":\"Monolith JSON projection\"}]}"},"rendered":{"html":null}}}`))
		case strings.Contains(body.Query, "AgentTaskResources"):
			_, _ = w.Write([]byte(`{"data":{"files":[],"links":[]}}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"getComponent":{"id":"component-a","componentType":"task","title":"Dashboard task","keyNumber":"RIID-5830"}}}`))
		}
	}))
	defer metadata.Close()
	t.Setenv(envTaskContextBaseURL, legacy.URL)
	t.Setenv("RIIDO_AI_SERVER_TASK_CONTEXT_GRAPHQL_URL", metadata.URL+"/graphql")
	reader, err := taskContextReaderFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	client := reader.(*riidoaiserver.AIAgentPrivateTaskContextClient)
	task, err := client.GetAgentDashboardTask(context.Background(), privateTaskContextRequest())
	if err != nil || task.Component.Title != "Dashboard task" {
		t.Fatalf("dashboard task=%+v error=%v", task, err)
	}
	normal, err := client.GetAIAgentTaskContextForRequest(context.Background(), privateTaskContextRequest())
	if err != nil || normal.Document.Content != "Monolith JSON projection" {
		t.Fatalf("legacy task context must remain available: %+v, %v", normal, err)
	}
}
