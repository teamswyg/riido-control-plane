package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

type fixedAuthorizationSchemeSelector struct {
	scheme AuthorizationScheme
	err    error
}

func (s fixedAuthorizationSchemeSelector) SelectAuthorizationScheme(string) (AuthorizationScheme, error) {
	return s.scheme, s.err
}

func TestAuthorizationSchemeRouterNeverDowngradesV2Failure(t *testing.T) {
	legacy := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "legacy"}}
	v2 := &countingAuthorizer{err: ErrAuthorizationForbidden}
	router, err := NewAuthorizationSchemeRouter(
		fixedAuthorizationSchemeSelector{scheme: AuthorizationSchemeAuthServiceV2}, legacy, v2,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = router.Authorize(context.Background(), "token", AuthorizationRequest{})
	if !errors.Is(err, ErrAuthorizationForbidden) || legacy.calls != 0 || v2.calls != 1 {
		t.Fatalf("err=%v legacy_calls=%d v2_calls=%d", err, legacy.calls, v2.calls)
	}
}

func TestAuthorizationSchemeRouterPreservesLegacyV1(t *testing.T) {
	legacy := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "legacy"}}
	v2 := &countingAuthorizer{err: ErrAuthorizationForbidden}
	router, _ := NewAuthorizationSchemeRouter(
		fixedAuthorizationSchemeSelector{scheme: AuthorizationSchemeLegacyV1}, legacy, v2,
	)
	result, err := router.Authorize(context.Background(), "token", AuthorizationRequest{})
	if err != nil || result.PrincipalID != "legacy" || legacy.calls != 1 || v2.calls != 0 {
		t.Fatalf("result=%+v err=%v legacy_calls=%d v2_calls=%d", result, err, legacy.calls, v2.calls)
	}
}

func TestAuthorizationSchemeRouterFailsClosedWhenV2Unavailable(t *testing.T) {
	legacy := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "legacy"}}
	router, _ := NewAuthorizationSchemeRouter(
		fixedAuthorizationSchemeSelector{scheme: AuthorizationSchemeAuthServiceV2}, legacy, nil,
	)
	_, err := router.Authorize(context.Background(), "token", AuthorizationRequest{})
	if !errors.Is(err, ErrAuthorizationSchemeUnavailable) || legacy.calls != 0 {
		t.Fatalf("err=%v legacy_calls=%d", err, legacy.calls)
	}
}

func TestAuthorizationSchemeRouterFailsClosedForUnknownScheme(t *testing.T) {
	legacy := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "legacy"}}
	router, _ := NewAuthorizationSchemeRouter(
		fixedAuthorizationSchemeSelector{scheme: AuthorizationScheme("future")}, legacy, nil,
	)
	_, err := router.Authorize(context.Background(), "token", AuthorizationRequest{})
	if !errors.Is(err, ErrAuthorizationSchemeUnavailable) || legacy.calls != 0 {
		t.Fatalf("err=%v legacy_calls=%d", err, legacy.calls)
	}
}
