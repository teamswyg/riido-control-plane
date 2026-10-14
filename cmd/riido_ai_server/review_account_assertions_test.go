package main

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func assertReviewTokenAuthorization(t *testing.T, authorizer riidoaiserver.RequestAuthorizer) {
	t.Helper()
	if authorizer == nil {
		t.Fatal("authorizer missing")
	}
	for _, req := range []riidoaiserver.AuthorizationRequest{
		{Resource: riidoaiserver.AuthorizationResourceAgentCatalog, Action: riidoaiserver.AuthorizationActionRead},
		{Resource: riidoaiserver.AuthorizationResourceAgent, Action: riidoaiserver.AuthorizationActionProviderStatusRead, AgentID: "store-review-agent"},
	} {
		if _, err := authorizer.Authorize(context.Background(), "review-token", req); err != nil {
			t.Fatalf("review token should authorize %+v: %v", req, err)
		}
	}
	_, err := authorizer.Authorize(context.Background(), "review-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionPoll,
		AgentID:  "store-review-agent",
	})
	if err == nil {
		t.Fatal("review token must not poll as daemon agent")
	}
}
