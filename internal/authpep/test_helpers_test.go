package authpep

import "context"

type countingAuthorizer struct {
	result AuthorizationResult
	err    error
	calls  int
}

func (a *countingAuthorizer) Authorize(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
	a.calls++
	if a.err != nil {
		return AuthorizationResult{}, a.err
	}
	return a.result, nil
}
