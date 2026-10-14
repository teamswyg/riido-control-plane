package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestCoalescingAuthorizerCollapsesConcurrentIdenticalRequests(t *testing.T) {
	next := newBlockingAuthorizer()
	authorizer, err := NewCoalescingAuthorizer(next)
	if err != nil {
		t.Fatalf("NewCoalescingAuthorizer: %v", err)
	}
	req := AuthorizationRequest{
		Resource:    AuthorizationResourceAIAgentClient,
		Action:      AuthorizationActionRead,
		WorkspaceID: "workspace-1",
		TaskID:      "task-1",
	}
	results := runConcurrentAuthorization(t, authorizer, "token-a", req, 32, next)
	for _, result := range results {
		if result.PrincipalID != "user-1" || result.WorkspaceID != "workspace-1" {
			t.Fatalf("result = %+v", result)
		}
	}
	if calls := next.calls.Load(); calls != 1 {
		t.Fatalf("underlying calls = %d, want 1", calls)
	}
}

func TestCoalescingAuthorizerKeepsDifferentRequestsSeparate(t *testing.T) {
	next := newBlockingAuthorizer()
	authorizer, err := NewCoalescingAuthorizer(next)
	if err != nil {
		t.Fatalf("NewCoalescingAuthorizer: %v", err)
	}
	start := make(chan struct{})
	errs := make(chan error, 2)
	for _, taskID := range []string{"task-a", "task-b"} {
		go func(taskID string) {
			<-start
			_, err := authorizer.Authorize(context.Background(), "token-a", AuthorizationRequest{
				Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead,
				WorkspaceID: "workspace-1", TaskID: taskID,
			})
			errs <- err
		}(taskID)
	}
	close(start)
	next.waitForCalls(t, 2)
	close(next.release)
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("Authorize: %v", err)
		}
	}
}

func TestCoalescingAuthorizerDoesNotExposeBearerTokenInKey(t *testing.T) {
	key := authorizationCoalescingKey("secret-token", AuthorizationRequest{Resource: AuthorizationResourceMetrics})
	if key == "" || strings.Contains(key, "secret-token") {
		t.Fatalf("coalescing key leaked bearer token: %q", key)
	}
}
