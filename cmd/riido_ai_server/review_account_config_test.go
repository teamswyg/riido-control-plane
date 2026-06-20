package main

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestConfigFromEnvIncludesReviewAccountProvisioning(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envReviewAccountTokenHash, testTokenSHA256("review-token"))
	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.ReviewProvision == nil {
		t.Fatal("review provisioning missing")
	}
	if config.ReviewProvision.Credential.Token != "" || config.ReviewProvision.Credential.TokenSHA256 == "" {
		t.Fatalf("review credential should use token hash only: %+v", config.ReviewProvision.Credential)
	}
	_, err = config.Authorizer.Authorize(context.Background(), "review-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgentCatalog,
		Action:   riidoaiserver.AuthorizationActionRead,
	})
	if err != nil {
		t.Fatalf("review token should read catalog: %v", err)
	}
}

func TestAuthorizerFromEnvIncludesReviewAccountCredential(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envReviewAccountTokenHash, testTokenSHA256("review-token"))
	authorizer, err := authorizerFromEnv()
	if err != nil {
		t.Fatalf("authorizerFromEnv: %v", err)
	}
	assertReviewTokenAuthorization(t, authorizer)
}
