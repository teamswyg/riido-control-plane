package main

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestParseAuthzTokensJSONRejectsUnknownField(t *testing.T) {
	_, err := parseAuthzTokensJSON(`[{"principal_id":"user-a","token":"static-token","scopes":["riido:*"],"extra":true}]`)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestAuthorizerFromEnvFallsBackFromStaticToExternalOnlyWhenUnauthenticated(t *testing.T) {
	clearRiidoAIServerEnv(t)
	harness := newExternalAuthorizerHarness(t)
	defer harness.Close()
	t.Setenv(envAuthzTokensJSON, `[{"principal_id":"static-user","token":"static-token","scopes":["metrics:read"]}]`)
	t.Setenv(envExternalAuthzURL, harness.URL)
	t.Setenv(envExternalAuthzAPIKey, "internal-key")
	t.Setenv(envExternalAuthzTimeout, "1")
	authorizer, err := authorizerFromEnv()
	if err != nil {
		t.Fatalf("authorizerFromEnv: %v", err)
	}
	assertStaticTokenDoesNotCallExternal(t, authorizer, harness)
	assertExternalTokenUsesFallback(t, authorizer, harness)
	assertForbiddenStaticScopeStopsFallback(t, authorizer, harness)
}

func TestAuthorizerFromEnvRejectsExternalAPIKeyWithoutEndpoint(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envExternalAuthzAPIKey, "internal-key")
	if _, err := authorizerFromEnv(); err == nil {
		t.Fatal("expected external api key without endpoint to fail")
	}
}

func TestAuthorizerFromEnvRejectsRemotePlainHTTPExternalAuthorizer(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envExternalAuthzURL, "http://authz.example.com/check")
	t.Setenv(envExternalAuthzAPIKey, "internal-key")
	if _, err := authorizerFromEnv(); err == nil ||
		!strings.Contains(err.Error(), envExternalAuthzURL) ||
		!strings.Contains(err.Error(), "https") {
		t.Fatalf("expected remote plain HTTP external authorizer rejection, got %v", err)
	}
}

func assertStaticTokenDoesNotCallExternal(t *testing.T, authorizer riidoaiserver.RequestAuthorizer, harness *externalAuthorizerHarness) {
	t.Helper()
	_, err := authorizer.Authorize(context.Background(), "static-token", metricsReadRequest())
	if err != nil {
		t.Fatalf("static authorize: %v", err)
	}
	if got := harness.Calls(); got != 0 {
		t.Fatalf("external calls after static token = %d", got)
	}
}
