package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestConfigFromEnvParsesPrivateTaskContextReader(t *testing.T) {
	clearRiidoAIServerEnv(t)
	var gotPaths []string
	var gotAuthorization []string
	var mu sync.Mutex
	taskContextServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotPaths = append(gotPaths, r.URL.String())
		gotAuthorization = append(gotAuthorization, r.Header.Get("Authorization"))
		mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
	wantPaths := []string{
		"/public/components/component-a/workspace",
		"/teams/team-a/components/component-a?getDocument=true",
		"/documents/providers/team-a/component-a?format=json",
		"/teams/team-a/components/component-a/files",
		"/teams/team-a/components/component-a/web-links",
	}
	sort.Strings(wantPaths)
	sort.Strings(gotPaths)
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("task context paths = %v", gotPaths)
	}
	for _, got := range gotAuthorization {
		if got != "Bearer user-jwt" {
			t.Fatalf("authorization = %q", got)
		}
	}
	if contextSnapshot.Component.Title != "Private task context from existing API server" ||
		contextSnapshot.Document.Content != "Provider projection from existing document API." {
		t.Fatalf("task context snapshot = %+v", contextSnapshot)
	}
}
