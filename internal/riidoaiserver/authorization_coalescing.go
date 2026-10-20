package riidoaiserver

import (
	"context"
	"errors"
	"sync"
)

type CoalescingAuthorizer struct {
	next  RequestAuthorizer
	mu    sync.Mutex
	calls map[string]*coalescedAuthorizationCall
}

type coalescedAuthorizationCall struct {
	done     chan struct{}
	result   AuthorizationResult
	err      error
	ctx      context.Context
	cancel   context.CancelFunc
	waiters  int
	finished bool
}

func NewCoalescingAuthorizer(next RequestAuthorizer) (*CoalescingAuthorizer, error) {
	if next == nil {
		return nil, errors.New("request authorizer is required")
	}
	return &CoalescingAuthorizer{
		next:  next,
		calls: map[string]*coalescedAuthorizationCall{},
	}, nil
}

func (a *CoalescingAuthorizer) Authorize(ctx context.Context, bearerToken string, req AuthorizationRequest) (AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthorizationResult{}, err
	}
	if a == nil || a.next == nil {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	key := authorizationCoalescingKey(bearerToken, req)
	call, owner := a.joinOrStart(ctx, key)
	defer a.release(key, call)
	if owner {
		go a.run(key, call, bearerToken, req)
	}
	return waitForCoalescedAuthorization(ctx, call)
}
