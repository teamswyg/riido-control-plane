package riidoaiserver

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type joinedAuthorizationContext struct {
	context.Context
	joined chan struct{}
	once   sync.Once
}

func (c *joinedAuthorizationContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.joined) })
	return c.Context.Done()
}

func TestCoalescingAuthorizerOwnerCancelDoesNotCancelLiveWaiter(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	next := authorizationFunc(func(ctx context.Context, _ string, _ AuthorizationRequest) (AuthorizationResult, error) {
		close(started)
		select {
		case <-ctx.Done():
			return AuthorizationResult{}, ctx.Err()
		case <-release:
			return AuthorizationResult{PrincipalID: "requester"}, nil
		}
	})
	a, _ := NewCoalescingAuthorizer(next)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	owner := make(chan error, 1)
	go func() { _, err := a.Authorize(ctx, "same-token", AuthorizationRequest{}); owner <- err }()
	<-started
	waitCtx := &joinedAuthorizationContext{Context: context.Background(), joined: make(chan struct{})}
	waiter := make(chan error, 1)
	go func() { _, err := a.Authorize(waitCtx, "same-token", AuthorizationRequest{}); waiter <- err }()
	<-waitCtx.joined
	cancel()
	if err := <-owner; !errors.Is(err, context.Canceled) {
		t.Fatalf("owner = %v, want cancellation", err)
	}
	close(release)
	select {
	case err := <-waiter:
		if err != nil {
			t.Fatalf("live identical request inherited cancelled owner: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("live waiter did not finish")
	}
}

type authorizationFunc func(context.Context, string, AuthorizationRequest) (AuthorizationResult, error)

func (f authorizationFunc) Authorize(ctx context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
	return f(ctx, token, req)
}
