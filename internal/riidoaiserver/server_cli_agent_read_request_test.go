package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIAgentReadRejectsAmbiguousRequests(t *testing.T) {
	const base = "/v1/cli/workspaces/workspace-fixture/tasks/task-fixture/agents"
	for _, tc := range []struct {
		name string
		edit func(*http.Request)
	}{
		{"method", func(r *http.Request) { r.Method = http.MethodPost }},
		{"origin", func(r *http.Request) { r.Header.Set("Origin", "https://app.example.test") }},
		{"upgrade", func(r *http.Request) { r.Header.Set("Upgrade", "websocket") }},
		{"duplicate bearer", func(r *http.Request) { r.Header.Add("Authorization", r.Header.Get("Authorization")) }},
		{"alternate header", func(r *http.Request) { r.Header.Set(aiAgentTokenHeader, "other") }},
		{"device", func(r *http.Request) { r.Header.Set(deviceIDHeader, "other") }},
		{"secret", func(r *http.Request) { r.Header.Set(deviceSecretHeader, "other") }},
		{"raw path", func(r *http.Request) { r.URL.RawPath = base }},
		{"query alias", func(r *http.Request) { r.URL.RawQuery = "cursor=01"; r.RequestURI += "?cursor=01" }},
		{"duplicate cursor", func(r *http.Request) { r.URL.RawQuery = "cursor=0&cursor=1"; r.RequestURI += "?cursor=0&cursor=1" }},
		{"unknown query", func(r *http.Request) { r.URL.RawQuery = "task=other"; r.RequestURI += "?task=other" }},
		{"negative", func(r *http.Request) { r.URL.RawQuery = "cursor=-1"; r.RequestURI += "?cursor=-1" }},
		{"URI mismatch", func(r *http.Request) { r.RequestURI = base + "/other" }},
		{"empty query", func(r *http.Request) { r.URL.ForceQuery = true; r.RequestURI += "?" }},
		{"thread ticket", func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+CLIThreadReadTokenPrefix+strings.Repeat("A", 43))
		}},
		{"app token", func(r *http.Request) { r.Header.Set("Authorization", "Bearer ordinary-app-token") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			a := authorizationFunc(func(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
				calls++
				return AuthorizationResult{PrincipalID: "user-1"}, nil
			})
			s := NewServer(ServerConfig{Authorizer: a, CLIAgentReadAuthorizer: a})
			r := httptest.NewRequest(http.MethodGet, base, nil)
			r.Header.Set("Authorization", "Bearer "+CLIAgentReadTokenPrefix+strings.Repeat("A", 43))
			tc.edit(r)
			w := httptest.NewRecorder()
			s.handleCLIAgentDirectory(w, r)
			if w.Code != http.StatusForbidden || calls != 0 {
				t.Fatal("unsafe request reached an authorizer", w.Code, calls)
			}
		})
	}
	s := NewServer(ServerConfig{})
	r := httptest.NewRequest(http.MethodGet, base, nil)
	r.Header.Set("Authorization", "Bearer "+CLIAgentReadTokenPrefix+strings.Repeat("A", 43))
	w := httptest.NewRecorder()
	s.handleCLIAgentDirectory(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatal("default-off reader was not unavailable", w.Code)
	}
}
