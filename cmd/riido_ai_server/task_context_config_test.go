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

func TestConfigFromEnvParsesTaskContextReader(t *testing.T) {
	clearRiidoAIServerEnv(t)
	var gotPath string
	var gotAPIKey string
	taskContextServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get(riidoaiserver.AIAgentTaskContextHeaderWorkspaceAPIKey)
		_ = json.NewEncoder(w).Encode(openAPITaskContextFixture())
	}))
	defer taskContextServer.Close()

	t.Setenv(envTaskContextBaseURL, taskContextServer.URL)
	t.Setenv(envTaskContextWorkspaceID, "workspace-a")
	t.Setenv(envTaskContextTeamID, "RIID")
	t.Setenv(envTaskContextAPIKey, "workspace-key")
	t.Setenv(envTaskContextTimeout, "1")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	contextSnapshot, err := config.TaskContextReader.GetAIAgentTaskContext(context.Background(), "component-a")
	if err != nil {
		t.Fatalf("GetAIAgentTaskContext: %v", err)
	}
	if gotPath != "/workspaces/workspace-a/open-api/v1/teams/RIID/components/component-a/ai-agent-context" || gotAPIKey != "workspace-key" {
		t.Fatalf("task context request path=%q apiKey=%q", gotPath, gotAPIKey)
	}
	if contextSnapshot.Component.BranchName != "RIID-4800-server-task-context-http-client-assignment-prompt-wiring" {
		t.Fatalf("task context snapshot = %+v", contextSnapshot)
	}
}

func TestConfigFromEnvRejectsPartialTaskContextConfig(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envTaskContextBaseURL, "https://api.riido.io")
	t.Setenv(envTaskContextWorkspaceID, "workspace-a")
	if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), "OpenAPI task context") {
		t.Fatalf("configFromEnv err=%v", err)
	}
}
