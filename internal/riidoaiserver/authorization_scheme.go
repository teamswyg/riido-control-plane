package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
)

type AuthorizationScheme string

const (
	AuthorizationSchemeLegacyV1      AuthorizationScheme = "legacy_v1"
	AuthorizationSchemeAuthServiceV2 AuthorizationScheme = "auth_service_v2"
)

var ErrAuthorizationSchemeUnavailable = errors.New("riidoaiserver: authorization scheme unavailable")

// AuthorizationSchemeSelector classifies credentials but never authenticates
// them. The selected provider remains solely responsible for verification.
type AuthorizationSchemeSelector interface {
	SelectAuthorizationScheme(string) (AuthorizationScheme, error)
}

type AuthorizationSchemeRouter struct {
	selector AuthorizationSchemeSelector
	legacy   RequestAuthorizer
	v2       RequestAuthorizer
}

func NewAuthorizationSchemeRouter(selector AuthorizationSchemeSelector, legacy, v2 RequestAuthorizer) (*AuthorizationSchemeRouter, error) {
	if selector == nil || legacy == nil {
		return nil, errors.New("authorization scheme selector and legacy_v1 authorizer are required")
	}
	return &AuthorizationSchemeRouter{selector: selector, legacy: legacy, v2: v2}, nil
}

func (r *AuthorizationSchemeRouter) Authorize(ctx context.Context, bearerToken string, req AuthorizationRequest) (AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthorizationResult{}, err
	}
	if r == nil || r.selector == nil || r.legacy == nil {
		return AuthorizationResult{}, ErrAuthorizationSchemeUnavailable
	}
	scheme, err := r.selector.SelectAuthorizationScheme(bearerToken)
	if err != nil {
		return AuthorizationResult{}, err
	}
	switch scheme {
	case AuthorizationSchemeLegacyV1:
		return r.legacy.Authorize(ctx, bearerToken, req)
	case AuthorizationSchemeAuthServiceV2:
		if r.v2 == nil {
			return AuthorizationResult{}, ErrAuthorizationSchemeUnavailable
		}
		return r.v2.Authorize(ctx, bearerToken, req)
	default:
		return AuthorizationResult{}, fmt.Errorf("%w: %q", ErrAuthorizationSchemeUnavailable, scheme)
	}
}
