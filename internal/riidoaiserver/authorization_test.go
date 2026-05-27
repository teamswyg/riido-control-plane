package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func TestStaticTokenAuthorizerAllowsScopedAgentToken(t *testing.T) {
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "daemon:jykim1",
		Token:       "agent-token",
		Scopes:      []string{"agent:jykim1:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	_, err = authorizer.Authorize(context.Background(), "agent-token", AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionPoll,
		AgentID:  "jykim1",
	})
	if err != nil {
		t.Fatalf("Authorize scoped agent: %v", err)
	}
	_, err = authorizer.Authorize(context.Background(), "agent-token", AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionPoll,
		AgentID:  "jykim2",
	})
	if !errors.Is(err, ErrAuthorizationForbidden) {
		t.Fatalf("Authorize other agent err=%v", err)
	}
}

func TestStaticTokenAuthorizerAllowsProviderStatusSyncScope(t *testing.T) {
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "daemon:jykim1",
		Token:       "agent-token",
		Scopes:      []string{"agent:jykim1:provider-status:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	_, err = authorizer.Authorize(context.Background(), "agent-token", AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionProviderStatusWrite,
		AgentID:  "jykim1",
	})
	if err != nil {
		t.Fatalf("Authorize provider status sync: %v", err)
	}
}

func TestStaticTokenAuthorizerSupportsSHA256TokenHash(t *testing.T) {
	sum := sha256.Sum256([]byte("viewer-token"))
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "viewer",
		TokenSHA256: hex.EncodeToString(sum[:]),
		Scopes:      []string{"component-task:task-a:events:read"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	_, err = authorizer.Authorize(context.Background(), "viewer-token", AuthorizationRequest{
		Resource: AuthorizationResourceComponentTaskEvents,
		Action:   AuthorizationActionEventsRead,
		TaskID:   "task-a",
	})
	if err != nil {
		t.Fatalf("Authorize hashed token: %v", err)
	}
}

func TestStaticTokenAuthorizerRejectsInvalidCredentials(t *testing.T) {
	if _, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "bad",
		Token:       "one",
		TokenSHA256: "two",
		Scopes:      []string{"riido:*"},
	}}); err == nil {
		t.Fatal("expected mutually exclusive token fields error")
	}
	if _, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "bad",
		TokenSHA256: "not-hex",
		Scopes:      []string{"riido:*"},
	}}); err == nil {
		t.Fatal("expected invalid hash error")
	}
	if _, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "bad",
		Token:       "secret",
	}}); err == nil {
		t.Fatal("expected missing scopes error")
	}
}
