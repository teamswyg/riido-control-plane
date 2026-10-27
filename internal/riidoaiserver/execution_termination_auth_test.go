package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

type terminationAuthorizer struct{ called bool }

func (a *terminationAuthorizer) Authorize(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
	a.called = true
	return AuthorizationResult{PrincipalID: "owner"}, nil
}

func TestExecutionTerminationDeviceAuthentication(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*http.Request)
		want   int
	}{
		{"bearer only", func(r *http.Request) { r.Header.Del(deviceIDHeader); r.Header.Del(deviceSecretHeader) }, 401},
		{"bad device and bearer", func(r *http.Request) { r.Header.Set(deviceSecretHeader, "bad") }, 401},
		{"missing lease", func(r *http.Request) { r.Header.Del(executionLeaseHeader) }, 401},
		{"duplicate lease", func(r *http.Request) { r.Header.Add(executionLeaseHeader, "other") }, 401},
		{"oversize lease", func(r *http.Request) { r.Header.Set(executionLeaseHeader, strings.Repeat("x", 8193)) }, 401},
		{"padded proof", func(r *http.Request) { r.Header.Set(deviceSecretHeader, " secret-sentinel") }, 401},
		{"wrong epoch", func(r *http.Request) { r.Header.Set(executionLeaseHeader, "other-epoch") }, 403},
		{"wrong assignment", func(r *http.Request) { r.URL.Path = "/v1/daemon/execution-claims/missing/termination" }, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newTerminationHTTPFixture(t, StoreConfig{})
			a := &terminationAuthorizer{}
			f.handler = NewServer(ServerConfig{Assignment: f.actor, AIAgentClient: f.devices, Authorizer: a}).Handler()
			r := f.request(terminationFixtureBody)
			r.Header.Set("Authorization", "Bearer valid-fixture")
			tc.change(r)
			w := serveTermination(f.handler, r)
			if w.Code != tc.want || a.called {
				t.Fatalf("status=%d bearer_called=%v", w.Code, a.called)
			}
		})
	}
}
