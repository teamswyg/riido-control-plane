package riidoaiserver

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestCoalescingAuthorizerPreservesValuesAndDoesNotCacheCompletedResults(t *testing.T) {
	type traceKey struct{}
	var calls atomic.Int64
	a, _ := NewCoalescingAuthorizer(authorizationFunc(func(ctx context.Context, _ string, _ AuthorizationRequest) (AuthorizationResult, error) {
		calls.Add(1)
		if ctx.Value(traceKey{}) != "trace-context" {
			t.Error("shared authorization lost request trace values")
		}
		return AuthorizationResult{PrincipalID: "requester"}, nil
	}))
	ctx := context.WithValue(context.Background(), traceKey{}, "trace-context")
	for range 2 {
		if _, err := a.Authorize(ctx, "token", AuthorizationRequest{}); err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 2 {
		t.Fatal("completed authorization was reused as a TTL cache")
	}
}

func TestCoalescingAuthorizationCancelledWaiterCannotConsumeReadySuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	close(done)
	call := &coalescedAuthorizationCall{done: done, result: AuthorizationResult{PrincipalID: "requester"}}
	for range 100 {
		result, err := waitForCoalescedAuthorization(ctx, call)
		if !errors.Is(err, context.Canceled) || result.PrincipalID != "" {
			t.Fatal("cancelled caller consumed a ready authorization")
		}
	}
}
