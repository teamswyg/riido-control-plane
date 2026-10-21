package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCoalescingAuthorizerLastWaiterCancelsAndLateFinishCannotEvictNewCall(t *testing.T) {
	var calls atomic.Int64
	started := make(chan context.Context, 2)
	first, second := make(chan struct{}), make(chan struct{})
	a, _ := NewCoalescingAuthorizer(authorizationFunc(func(ctx context.Context, _ string, _ AuthorizationRequest) (AuthorizationResult, error) {
		n := calls.Add(1)
		started <- ctx
		if n == 1 {
			<-first // Deliberately emulate a dependency that returns late after cancellation.
		} else {
			<-second
		}
		return AuthorizationResult{PrincipalID: "requester"}, nil
	}))
	ctx, cancel := context.WithCancel(context.Background())
	retired := make(chan error, 1)
	go func() { _, err := a.Authorize(ctx, "token", AuthorizationRequest{}); retired <- err }()
	shared := <-started
	key := authorizationCoalescingKey("token", AuthorizationRequest{})
	a.mu.Lock()
	old := a.calls[key]
	a.mu.Unlock()
	cancel()
	if err := <-retired; !errors.Is(err, context.Canceled) || shared.Err() != context.Canceled {
		t.Fatal("last waiter did not retire and cancel the shared hop")
	}
	current := make(chan error, 1)
	go func() { _, err := a.Authorize(context.Background(), "token", AuthorizationRequest{}); current <- err }()
	<-started
	close(first)
	<-old.done
	a.mu.Lock()
	live := a.calls[key]
	a.mu.Unlock()
	if live == nil || live == old {
		t.Fatal("late old completion removed the new in-flight request")
	}
	close(second)
	if err := <-current; err != nil || calls.Load() != 2 {
		t.Fatalf("fresh authorization failed or repeated: %v, calls=%d", err, calls.Load())
	}
}

func TestCoalescingAuthorizerPanicReleasesWaitersWithoutDetails(t *testing.T) {
	a, _ := NewCoalescingAuthorizer(authorizationFunc(func(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
		panic("PRIVATE_CREDENTIAL_CANARY")
	}))
	result, err := a.Authorize(context.Background(), "token", AuthorizationRequest{})
	if err == nil || result.PrincipalID != "" || strings.Contains(err.Error(), "PRIVATE_CREDENTIAL_CANARY") {
		t.Fatal("dependency panic escaped or allowed authorization")
	}
	if len(a.calls) != 0 {
		t.Fatal("failed dependency left an in-flight authorization cached")
	}
}
