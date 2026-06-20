package main

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func assertExternalTokenUsesFallback(t *testing.T, authorizer riidoaiserver.RequestAuthorizer, harness *externalAuthorizerHarness) {
	t.Helper()
	_, err := authorizer.Authorize(context.Background(), "external-token", metricsReadRequest())
	if err != nil {
		t.Fatalf("external authorize: %v", err)
	}
	if got := harness.Calls(); got != 1 {
		t.Fatalf("external calls after fallback = %d", got)
	}
}

func assertForbiddenStaticScopeStopsFallback(t *testing.T, authorizer riidoaiserver.RequestAuthorizer, harness *externalAuthorizerHarness) {
	t.Helper()
	_, err := authorizer.Authorize(context.Background(), "static-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionPoll,
		AgentID:  "agent-a",
	})
	if err == nil {
		t.Fatal("expected forbidden static scope to stop fallback")
	}
	if got := harness.Calls(); got != 1 {
		t.Fatalf("external should not run after forbidden static scope, calls=%d", got)
	}
}

func metricsReadRequest() riidoaiserver.AuthorizationRequest {
	return riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceMetrics,
		Action:   riidoaiserver.AuthorizationActionRead,
	}
}
