package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestCloneAuthorizationResultCopiesRoles(t *testing.T) {
	original := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: "workspace-1",
		Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
	}
	cloned := cloneAuthorizationResult(original)
	cloned.Roles[0] = AgentCatalogRole("member")
	if original.Roles[0] != AgentCatalogRoleAdmin {
		t.Fatalf("clone mutated original roles: original=%+v cloned=%+v", original, cloned)
	}
	empty := cloneAuthorizationResult(AuthorizationResult{PrincipalID: "user-2"})
	if empty.PrincipalID != "user-2" || len(empty.Roles) != 0 {
		t.Fatalf("empty clone = %+v", empty)
	}
}

func TestWaitForCoalescedAuthorizationContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := waitForCoalescedAuthorization(ctx, &coalescedAuthorizationCall{
		done: make(chan struct{}),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("wait err = %v, want context.Canceled", err)
	}
}
