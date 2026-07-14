package main

import (
	"context"
	"encoding/base64"
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

func TestAuthorizerFromEnvRejectsPartialAuthProfile(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAuthIssuer, "https://auth.riido.io")
	if _, err := authorizerFromEnv(); err == nil || !strings.Contains(err.Error(), envAuthIssuer) {
		t.Fatalf("partial Auth profile err=%v", err)
	}
}

func TestAuthorizerFromEnvRequiresConsumerDomainPDPForAuthProfile(t *testing.T) {
	clearRiidoAIServerEnv(t)
	setAuthProfileEnv(t)
	if _, err := authorizerFromEnv(); err == nil || !strings.Contains(err.Error(), envExternalAuthzURL) {
		t.Fatalf("missing domain PDP err=%v", err)
	}
}

func TestAuthorizerFromEnvComposesAuthPEPWithoutChangingStaticTokenBehavior(t *testing.T) {
	clearRiidoAIServerEnv(t)
	harness := newExternalAuthorizerHarness(t)
	defer harness.Close()
	setAuthProfileEnv(t)
	t.Setenv(envExternalAuthzURL, harness.URL)
	t.Setenv(envExternalAuthzAPIKey, "internal-key")
	t.Setenv(envExternalAuthzTimeout, "1")
	t.Setenv(envAuthzTokensJSON, `[{"principal_id":"static-user","token":"static-token","scopes":["metrics:read"]}]`)
	authorizer, err := authorizerFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	result, err := authorizer.Authorize(context.Background(), "static-token", metricsReadRequest())
	if err != nil || result.PrincipalID != "static-user" || harness.Calls() != 0 {
		t.Fatalf("result=%+v err=%v external_calls=%d", result, err, harness.Calls())
	}
}

func TestAuthorizerFromEnvKeepsForeignIssuerJWTOnLegacyAuthorizer(t *testing.T) {
	clearRiidoAIServerEnv(t)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"legacy"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"https://legacy-idp.example","sub":"legacy-user"}`))
	legacyJWT := header + "." + payload + ".legacy-signature"
	harness := newExternalAuthorizerHarnessForToken(t, legacyJWT)
	defer harness.Close()
	setAuthProfileEnv(t)
	t.Setenv(envExternalAuthzURL, harness.URL)
	t.Setenv(envExternalAuthzAPIKey, "internal-key")
	t.Setenv(envExternalAuthzTimeout, "1")
	authorizer, err := authorizerFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	result, err := authorizer.Authorize(context.Background(), legacyJWT, metricsReadRequest())
	if err != nil || result.PrincipalID != "external-user" || harness.Calls() != 1 {
		t.Fatalf("result=%+v err=%v external_calls=%d", result, err, harness.Calls())
	}
}

func TestAuthorizerFromEnvRejectsUnboundedAuthHTTPTimeout(t *testing.T) {
	clearRiidoAIServerEnv(t)
	harness := newExternalAuthorizerHarness(t)
	defer harness.Close()
	setAuthProfileEnv(t)
	t.Setenv(envExternalAuthzURL, harness.URL)
	t.Setenv(envAuthHTTPTimeout, "11")
	if _, err := authorizerFromEnv(); err == nil || !strings.Contains(err.Error(), envAuthHTTPTimeout) {
		t.Fatalf("unbounded Auth timeout err=%v", err)
	}
}

func setAuthProfileEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envAuthIssuer, "https://auth.riido.io")
	t.Setenv(envAuthResource, "https://ai-api.riido.io")
	t.Setenv(envAuthAuthorizationProfile, "riido-control-plane.production.v1")
	t.Setenv(envAuthIntrospectionClientID, "riido-control-plane")
	t.Setenv(envAuthIntrospectionClientSecret, "control-plane-introspection-secret-material")
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
