package authpep

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"
	"time"
)

func TestJWTAccessTokenAuthorizerNeverDowngradesConfiguredIssuer(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	claims := validJWTClaims(now)
	for name, token := range map[string]string{
		"invalid signature": jwtShapedTokenForTest(t, `{"iss":"https://auth.riido.io"}`),
		"duplicate issuer":  jwtShapedTokenForTest(t, `{"iss":"https://legacy-idp.example","iss":"https://auth.riido.io"}`),
	} {
		t.Run(name, func(t *testing.T) {
			jwt := newJWTAuthorizerForTest(t, now, &key.PublicKey, fixedAccessTokenStatusResolver{status: statusFromJWTClaims(claims)}, &countingAuthorizer{})
			fallback := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "legacy"}}
			chain, _ := NewFallbackAuthorizer(jwt, fallback)
			_, err := chain.Authorize(context.Background(), token, AuthorizationRequest{Resource: AuthorizationResourceMetrics, Action: AuthorizationActionRead})
			if !errors.Is(err, ErrAuthorizationInvalidCredential) || fallback.calls != 0 {
				t.Fatalf("err=%v fallback_calls=%d", err, fallback.calls)
			}
		})
	}
}

func TestDecodeExactJWTJSONRejectsDuplicateKeys(t *testing.T) {
	for name, raw := range map[string]string{
		"header algorithm": `{"alg":"ES256","alg":"none","kid":"auth-key-202607","typ":"at+jwt"}`,
		"claims issuer":    `{"iss":"https://legacy-idp.example","iss":"https://auth.riido.io"}`,
		"claims audience":  `{"aud":"https://ai-api.riido.io","aud":"https://work.riido.io"}`,
	} {
		t.Run(name, func(t *testing.T) {
			var decoded map[string]any
			if err := decodeExactJWTJSON([]byte(raw), &decoded); err == nil {
				t.Fatal("expected duplicate JWT JSON key to be rejected")
			}
		})
	}
}
