package authpep

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLegacyApplicationTokensRejectCLIWithoutCallingPDP(t *testing.T) {
	cases := map[string]string{
		"current":                 `{"clientId":"riido-cli","cliGrantId":"grant"}`,
		"client only":             `{"clientId":"riido-cli"}`,
		"grant only":              `{"cliGrantId":"grant"}`,
		"null grant":              `{"cliGrantId":null}`,
		"invalid grant":           `{"cliGrantId":{}}`,
		"case variant":            `{"CLIGRANTID":"grant"}`,
		"escaped key":             `{"cli\u0047rantId":"grant"}`,
		"escaped client":          `{"clientId":"riido-\u0063li"}`,
		"duplicate client":        `{"clientId":"riido-cli","clientId":"browser"}`,
		"duplicate inverse":       `{"clientId":"browser","clientId":"riido-cli"}`,
		"marker before malformed": `{"cliGrantId":"grant",`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			next := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "synthetic-user"}}
			authz := RestrictLegacyApplicationTokens(next)
			_, err := authz.Authorize(context.Background(), jwtShapedTokenForTest(t, payload), AuthorizationRequest{})
			if !errors.Is(err, ErrAuthorizationForbidden) || next.calls != 0 {
				t.Fatalf("err=%v calls=%d", err, next.calls)
			}
		})
	}
}

func TestLegacyApplicationTokensPreserveAppAndOpaqueCredentials(t *testing.T) {
	for name, raw := range map[string]string{
		"opaque": "opaque-static-token", "missing": "",
		"app":           jwtShapedTokenForTest(t, `{"id":"person","iss":"https://legacy.example"}`),
		"other client":  jwtShapedTokenForTest(t, `{"clientId":"another-client"}`),
		"nested marker": jwtShapedTokenForTest(t, `{"profile":{"cliGrantId":"other"}}`),
		"malformed":     "header.!payload.signature",
	} {
		t.Run(name, func(t *testing.T) {
			next := &countingAuthorizer{err: ErrAuthorizationUnauthenticated}
			_, err := RestrictLegacyApplicationTokens(next).Authorize(context.Background(), raw, AuthorizationRequest{})
			if !errors.Is(err, ErrAuthorizationUnauthenticated) || next.calls != 1 {
				t.Fatalf("err=%v calls=%d", err, next.calls)
			}
		})
	}
}

func TestLegacyApplicationTokensRejectOversizeJWTWithoutFallback(t *testing.T) {
	next := &countingAuthorizer{}
	raw := jwtShapedTokenForTest(t, `{"padding":"`+strings.Repeat("x", maxJWTAccessTokenBytes)+`","cliGrantId":"grant"}`)
	_, err := RestrictLegacyApplicationTokens(next).Authorize(context.Background(), raw, AuthorizationRequest{})
	if !errors.Is(err, ErrAuthorizationForbidden) || next.calls != 0 {
		t.Fatalf("err=%v calls=%d", err, next.calls)
	}
}

func TestLegacyApplicationTokensPreserveAbsentAndCancelledAuthorizer(t *testing.T) {
	if RestrictLegacyApplicationTokens(nil) != nil {
		t.Fatal("absent authorizer was enabled")
	}
	next := &countingAuthorizer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := RestrictLegacyApplicationTokens(next).Authorize(ctx, "opaque", AuthorizationRequest{})
	if !errors.Is(err, context.Canceled) || next.calls != 0 {
		t.Fatalf("err=%v calls=%d", err, next.calls)
	}
}
