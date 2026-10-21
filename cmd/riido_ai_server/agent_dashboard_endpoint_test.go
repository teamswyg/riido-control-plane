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
		if strings.Contains(body.Query, "AgentDashboardLocation") {
			_, _ = w.Write([]byte(`{"data":{"getComponentWorkspace":{"id":"component-a","componentType":"task","team":{"id":"team-a","workspace":{"id":"workspace-a"}}}}}`))
		} else {
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
	if err != nil || normal.Document.Content != "<p>Provider document from Mongo.</p>" {
		t.Fatalf("legacy task context must remain available: %+v, %v", normal, err)
	}
}
