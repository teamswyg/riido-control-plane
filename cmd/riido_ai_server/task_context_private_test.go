package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestConfigFromEnvParsesPrivateTaskContextReader(t *testing.T) {
	clearRiidoAIServerEnv(t)
	var gotPaths []string
	var gotAuthorization []string
	taskContextServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.String())
		gotAuthorization = append(gotAuthorization, r.Header.Get("Authorization"))
		writePrivateTaskContextResponse(w, r)
	}))
	defer taskContextServer.Close()

	t.Setenv(envTaskContextBaseURL, taskContextServer.URL)
	t.Setenv(envTaskContextTimeout, "1")
	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	requestReader, ok := config.TaskContextReader.(riidoaiserver.AIAgentTaskContextRequestReader)
	if !ok {
		t.Fatalf("task context reader should support request-scoped JWT")
	}
	contextSnapshot, err := requestReader.GetAIAgentTaskContextForRequest(context.Background(), privateTaskContextRequest())
	if err != nil {
		t.Fatalf("GetAIAgentTaskContextForRequest: %v", err)
	}
	if !reflect.DeepEqual(gotPaths, []string{"/public/components/component-a/workspace", "/teams/team-a/components/component-a?getDocument=true"}) {
		t.Fatalf("task context paths = %v", gotPaths)
	}
	for _, got := range gotAuthorization {
		if got != "Bearer user-jwt" {
			t.Fatalf("authorization = %q", got)
		}
	}
	if contextSnapshot.Component.Title != "Private task context from existing API server" ||
		contextSnapshot.Document.Content != "<p>Existing API server private document.</p>" {
		t.Fatalf("task context snapshot = %+v", contextSnapshot)
	}
}
