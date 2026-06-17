package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExternalHTTPAuthorizerAllowsScopedRequest(t *testing.T) {
	var got externalAuthorizerRequest
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q", got)
		}
		if got := r.Header.Get(ExternalAuthorizerAPIKeyHeader); got != "internal-key" {
			t.Fatalf("api key header = %q", got)
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(externalAuthorizerResponse{
			SchemaVersion: ExternalAuthorizerResponseSchemaVersion,
			Allowed:       true,
			PrincipalID:   "admin-1",
			Roles:         []AgentCatalogRole{AgentCatalogRoleAdmin},
		})
	}))
	defer authorizerServer.Close()

	authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{
		Endpoint: authorizerServer.URL,
		Audience: "riido-api",
		APIKey:   " internal-key ",
	})
	if err != nil {
		t.Fatalf("NewExternalHTTPAuthorizer: %v", err)
	}
	result, err := authorizer.Authorize(context.Background(), "external-token", AuthorizationRequest{
		Resource: AuthorizationResourceAgentCatalog,
		Action:   AuthorizationActionUpdate,
		AgentID:  "agent-a",
	})
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if result.PrincipalID != "admin-1" || len(result.Roles) != 1 || result.Roles[0] != AgentCatalogRoleAdmin {
		t.Fatalf("result = %+v", result)
	}
	if got.SchemaVersion != ExternalAuthorizerRequestSchemaVersion ||
		got.BearerToken != "external-token" ||
		got.Audience != "riido-api" ||
		got.Request.Resource != AuthorizationResourceAgentCatalog ||
		got.Request.Action != AuthorizationActionUpdate ||
		got.Request.AgentID != "agent-a" {
		t.Fatalf("request = %+v", got)
	}
}

func TestExternalHTTPAuthorizerMapsDeniedResponses(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{name: "unauthenticated", status: http.StatusUnauthorized, body: `{}`, wantErr: ErrAuthorizationUnauthenticated},
		{name: "forbidden", status: http.StatusForbidden, body: `{}`, wantErr: ErrAuthorizationForbidden},
		{name: "allowed false", status: http.StatusOK, body: `{"schema_version":"riido-external-authorizer-response.v1","allowed":false}`, wantErr: ErrAuthorizationForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer authorizerServer.Close()
			authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: authorizerServer.URL})
			if err != nil {
				t.Fatalf("NewExternalHTTPAuthorizer: %v", err)
			}
			_, err = authorizer.Authorize(context.Background(), "external-token", AuthorizationRequest{
				Resource: AuthorizationResourceMetrics,
				Action:   AuthorizationActionRead,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Authorize err=%v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestExternalHTTPAuthorizerFailsClosedOnProviderError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "upstream error", status: http.StatusInternalServerError, body: `{}`},
		{name: "malformed json", status: http.StatusOK, body: `{`},
		{name: "unsupported schema", status: http.StatusOK, body: `{"schema_version":"wrong","allowed":true,"principal_id":"user-1"}`},
		{name: "missing principal", status: http.StatusOK, body: `{"schema_version":"riido-external-authorizer-response.v1","allowed":true}`},
		{name: "invalid role", status: http.StatusOK, body: `{"schema_version":"riido-external-authorizer-response.v1","allowed":true,"principal_id":"user-1","roles":["owner"]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer authorizerServer.Close()
			authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: authorizerServer.URL})
			if err != nil {
				t.Fatalf("NewExternalHTTPAuthorizer: %v", err)
			}
			_, err = authorizer.Authorize(context.Background(), "external-token", AuthorizationRequest{
				Resource: AuthorizationResourceMetrics,
				Action:   AuthorizationActionRead,
			})
			if err == nil || errors.Is(err, ErrAuthorizationUnauthenticated) || errors.Is(err, ErrAuthorizationForbidden) {
				t.Fatalf("Authorize err=%v, want fail-closed service error", err)
			}
		})
	}
}

func TestExternalHTTPAuthorizerRejectsUnsafeEndpoint(t *testing.T) {
	for _, endpoint := range []string{
		"",
		"ftp://authz.example.com",
		"https://user:pass@authz.example.com",
		"https://authz.example.com/check?token=raw",
		"https://authz.example.com/check#fragment",
		"http://authz.example.com/check",
		"http://192.0.2.10/check",
	} {
		t.Run(strings.ReplaceAll(endpoint, "/", "_"), func(t *testing.T) {
			if _, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: endpoint}); err == nil {
				t.Fatalf("expected endpoint %q to be rejected", endpoint)
			}
		})
	}
}

func TestExternalHTTPAuthorizerAllowsPlainHTTPOnlyForLoopback(t *testing.T) {
	for _, endpoint := range []string{
		"http://localhost/check",
		"http://127.0.0.1:8080/check",
		"http://[::1]:8080/check",
	} {
		t.Run(endpoint, func(t *testing.T) {
			if _, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: endpoint}); err != nil {
				t.Fatalf("expected loopback endpoint %q to be allowed: %v", endpoint, err)
			}
		})
	}
}

func TestExternalHTTPAuthorizerRejectsUnsafeAPIKey(t *testing.T) {
	if _, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{
		Endpoint: "https://authz.example.com/check",
		APIKey:   "key\nother: value",
	}); err == nil {
		t.Fatal("expected unsafe api key to be rejected")
	}
}

func TestFallbackAuthorizerUsesNextOnlyForUnauthenticated(t *testing.T) {
	static, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "static-user",
		Token:       "static-token",
		Scopes:      []string{"metrics:read"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	next := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "external-user"}}
	chain, err := NewFallbackAuthorizer(static, next)
	if err != nil {
		t.Fatalf("NewFallbackAuthorizer: %v", err)
	}

	result, err := chain.Authorize(context.Background(), "external-token", AuthorizationRequest{
		Resource: AuthorizationResourceMetrics,
		Action:   AuthorizationActionRead,
	})
	if err != nil {
		t.Fatalf("Authorize fallback token: %v", err)
	}
	if result.PrincipalID != "external-user" || next.calls != 1 {
		t.Fatalf("fallback result=%+v calls=%d", result, next.calls)
	}

	_, err = chain.Authorize(context.Background(), "static-token", AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionPoll,
		AgentID:  "agent-a",
	})
	if !errors.Is(err, ErrAuthorizationForbidden) {
		t.Fatalf("Authorize forbidden static token err=%v", err)
	}
	if next.calls != 1 {
		t.Fatalf("fallback should not run after forbidden, calls=%d", next.calls)
	}
}

type countingAuthorizer struct {
	result AuthorizationResult
	err    error
	calls  int
}

func (a *countingAuthorizer) Authorize(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
	a.calls++
	if a.err != nil {
		return AuthorizationResult{}, a.err
	}
	return a.result, nil
}
