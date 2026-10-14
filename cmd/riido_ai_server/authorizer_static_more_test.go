package main

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestParseAuthzTokensJSONBlankAndValid(t *testing.T) {
	if authz, err := parseAuthzTokensJSON("  "); err != nil || authz != nil {
		t.Fatalf("blank authz = %T, %v; want nil, nil", authz, err)
	}
	authz, err := parseAuthzTokensJSON(`[
		{"principal_id":"user-a","token":"token-a","scopes":["metrics:read"]}
	]`)
	if err != nil || authz == nil {
		t.Fatalf("valid authz = %T, %v; want authorizer", authz, err)
	}
	if _, err := authz.Authorize(context.Background(), "token-a", metricsReadRequest()); err != nil {
		t.Fatalf("valid token authorize: %v", err)
	}
}

func TestStaticAuthorizerFromEnvUsesReviewCredentialOnly(t *testing.T) {
	clearRiidoAIServerEnv(t)
	credential := riidoaiserver.StaticTokenCredential{
		PrincipalID: "review-user",
		TokenSHA256: testTokenSHA256("review-token"),
		Scopes:      []string{"agent-catalog:read"},
	}
	authz, err := staticAuthorizerFromEnv(&riidoaiserver.ReviewAccountProvisioning{
		Credential: credential,
	})
	if err != nil || authz == nil {
		t.Fatalf("staticAuthorizerFromEnv = %T, %v; want authorizer", authz, err)
	}
	_, err = authz.Authorize(context.Background(), "review-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgentCatalog,
		Action:   riidoaiserver.AuthorizationActionRead,
	})
	if err != nil {
		t.Fatalf("review credential authorize: %v", err)
	}
}
