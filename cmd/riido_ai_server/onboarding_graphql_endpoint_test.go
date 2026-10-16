package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestOnboardingProvenanceUsesSeparateGraphQLEndpoint(t *testing.T) {
	clearRiidoAIServerEnv(t)
	var restCalls, graphqlCalls atomic.Int64
	rest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restCalls.Add(1)
		http.NotFound(w, r)
	}))
	defer rest.Close()
	metadata := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		graphqlCalls.Add(1)
		if r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer user-jwt" {
			t.Error("metadata request lost endpoint or viewer authorization")
		}
		var body struct{ Query string }
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("invalid GraphQL request")
		}
		switch {
		case strings.Contains(body.Query, "OnboardingTaskLocation"):
			_, _ = w.Write([]byte(`{"data":{"getComponentWorkspace":{"id":"component-a","componentType":"task","team":{"id":"team-a","workspace":{"id":"workspace-a"}}}}}`))
		case strings.Contains(body.Query, "OnboardingFirstTaskBundleTemplate"):
			_, _ = w.Write([]byte(`{"data":{"getMyAccountInfo":{"id":"user-a"},"onboardingFirstTaskBundleTemplate":"schedule"}}`))
		default:
			t.Error("unexpected metadata operation")
		}
	}))
	defer metadata.Close()
	t.Setenv(envTaskContextBaseURL, rest.URL)
	t.Setenv(envTaskContextGraphQLURL, metadata.URL+"/graphql")
	reader, err := taskContextReaderFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	client := reader.(*riidoaiserver.AIAgentPrivateTaskContextClient)
	answer, err := client.OnboardingFirstTaskBundleTemplate(context.Background(), "user-jwt", "workspace-a", "component-a")
	if err != nil || answer.ViewerID != "user-a" || answer.Template != riidoaiserver.AIAgentOnboardingTemplateSchedule {
		t.Fatal("onboarding provenance did not use the configured GraphQL server")
	}
	if restCalls.Load() != 0 || graphqlCalls.Load() != 2 {
		t.Fatal("onboarding provenance crossed into REST or repeated metadata reads")
	}
}
