package authpep

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type staticJWTKeyProvider struct {
	key *ecdsa.PublicKey
	err error
}

func (p staticJWTKeyProvider) VerificationKey(context.Context, string) (*ecdsa.PublicKey, error) {
	return p.key, p.err
}

type fixedAccessTokenStatusResolver struct {
	status AccessTokenStatus
	err    error
}

func (r fixedAccessTokenStatusResolver) ResolveAccessToken(context.Context, string) (AccessTokenStatus, error) {
	return r.status, r.err
}

func TestJWTAccessTokenAuthorizerRequiresAuthStatusExactScopeAndDomainDecision(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	claims := validJWTClaims(now)
	token := signJWTForTest(t, key, claims)
	status := statusFromJWTClaims(claims)
	domain := &countingAuthorizer{result: AuthorizationResult{PrincipalID: claims.Subject, WorkspaceID: "workspace-1"}}
	authorizer := newJWTAuthorizerForTest(t, now, &key.PublicKey, fixedAccessTokenStatusResolver{status: status}, domain)

	result, err := authorizer.Authorize(context.Background(), token, AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, WorkspaceID: "workspace-1",
	})
	if err != nil || result.PrincipalID != claims.Subject || domain.calls != 1 {
		t.Fatalf("result=%+v err=%v domain_calls=%d", result, err, domain.calls)
	}

	_, err = authorizer.Authorize(context.Background(), token, AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionUpdate, WorkspaceID: "workspace-1",
	})
	if !errors.Is(err, ErrAuthorizationForbidden) || domain.calls != 1 {
		t.Fatalf("missing exact write scope err=%v domain_calls=%d", err, domain.calls)
	}
}

func TestJWTAccessTokenAuthorizerRejectsSubstitutionAndStopsFallback(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	claims := validJWTClaims(now)
	validStatus := statusFromJWTClaims(claims)
	cases := map[string]struct {
		mutate func(*jwtAccessClaims)
		status AccessTokenStatus
	}{
		"wrong audience":         {mutate: func(value *jwtAccessClaims) { value.Audience = "https://work.riido.io" }, status: validStatus},
		"wrong profile":          {mutate: func(value *jwtAccessClaims) { value.AuthorizationProfile = "riido-ci.production.v1" }, status: validStatus},
		"wildcard scope":         {mutate: func(value *jwtAccessClaims) { value.Scope = "ai-agent:* email openid" }, status: validStatus},
		"email newline":          {mutate: func(value *jwtAccessClaims) { value.Email = "jy\nkim@swyg.im" }, status: validStatus},
		"introspection mismatch": {mutate: func(*jwtAccessClaims) {}, status: func() AccessTokenStatus { value := validStatus; value.Subject = "human:other@swyg.im"; return value }()},
		"inactive":               {mutate: func(*jwtAccessClaims) {}, status: AccessTokenStatus{}},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			mutated := claims
			test.mutate(&mutated)
			token := signJWTForTest(t, key, mutated)
			next := &countingAuthorizer{result: AuthorizationResult{PrincipalID: mutated.Subject, WorkspaceID: "workspace-1"}}
			jwt := newJWTAuthorizerForTest(t, now, &key.PublicKey, fixedAccessTokenStatusResolver{status: test.status}, next)
			fallback := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "fallback"}}
			chain, err := NewFallbackAuthorizer(jwt, fallback)
			if err != nil {
				t.Fatal(err)
			}
			_, err = chain.Authorize(context.Background(), token, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, WorkspaceID: "workspace-1"})
			if !errors.Is(err, ErrAuthorizationInvalidCredential) || fallback.calls != 0 {
				t.Fatalf("err=%v fallback_calls=%d", err, fallback.calls)
			}
		})
	}
}

func TestJWTAccessTokenAuthorizerKeepsLegacyOpaqueAndForeignIssuerJWTPaths(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	claims := validJWTClaims(now)
	for name, token := range map[string]string{
		"opaque":         "legacy-static-token",
		"foreign issuer": jwtShapedTokenForTest(t, `{"iss":"https://legacy-idp.example"}`),
		"missing issuer": jwtShapedTokenForTest(t, `{"sub":"legacy-user"}`),
	} {
		t.Run(name, func(t *testing.T) {
			jwt := newJWTAuthorizerForTest(t, now, &key.PublicKey, fixedAccessTokenStatusResolver{status: statusFromJWTClaims(claims)}, &countingAuthorizer{})
			fallback := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "legacy"}}
			chain, _ := NewFallbackAuthorizer(jwt, fallback)
			result, err := chain.Authorize(context.Background(), token, AuthorizationRequest{Resource: AuthorizationResourceMetrics, Action: AuthorizationActionRead})
			if err != nil || result.PrincipalID != "legacy" || fallback.calls != 1 {
				t.Fatalf("result=%+v err=%v fallback_calls=%d", result, err, fallback.calls)
			}
		})
	}
}

func TestJWTAccessTokenAuthorizerRejectsDomainPrincipalOrWorkspaceSubstitution(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	claims := validJWTClaims(now)
	token := signJWTForTest(t, key, claims)
	for name, result := range map[string]AuthorizationResult{
		"principal": {PrincipalID: "human:other@swyg.im", WorkspaceID: "workspace-1"},
		"workspace": {PrincipalID: claims.Subject, WorkspaceID: "workspace-2"},
	} {
		t.Run(name, func(t *testing.T) {
			authorizer := newJWTAuthorizerForTest(t, now, &key.PublicKey, fixedAccessTokenStatusResolver{status: statusFromJWTClaims(claims)}, &countingAuthorizer{result: result})
			_, err := authorizer.Authorize(context.Background(), token, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, WorkspaceID: "workspace-1"})
			if !errors.Is(err, ErrAuthorizationForbidden) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func newJWTAuthorizerForTest(t *testing.T, now time.Time, key *ecdsa.PublicKey, status AccessTokenStatusResolver, domain RequestAuthorizer) *JWTAccessTokenAuthorizer {
	t.Helper()
	authorizer, err := NewJWTAccessTokenAuthorizer(JWTAuthorizationConfig{
		Issuer: "https://auth.riido.io", Audience: "https://ai-api.riido.io", AuthorizationProfile: "riido-control-plane.production.v1",
		ClockSkew: 30 * time.Second, Clock: func() time.Time { return now },
	}, staticJWTKeyProvider{key: key}, status, domain)
	if err != nil {
		t.Fatal(err)
	}
	return authorizer
}

func validJWTClaims(now time.Time) jwtAccessClaims {
	return jwtAccessClaims{
		Issuer: "https://auth.riido.io", Subject: "human:jykim@swyg.im", Audience: "https://ai-api.riido.io",
		ExpiresAt: now.Add(5 * time.Minute).Unix(), IssuedAt: now.Unix(), NotBefore: now.Unix(), JWTID: "token-id-12345678",
		ClientID: "riido-interaction", Scope: "ai-agent:read email openid", Email: "jykim@swyg.im", EmailVerified: true,
		AuthorizationProfile: "riido-control-plane.production.v1",
	}
}

func statusFromJWTClaims(claims jwtAccessClaims) AccessTokenStatus {
	return AccessTokenStatus{
		Active: true, Scope: claims.Scope, ClientID: claims.ClientID, Subject: claims.Subject, TokenType: "Bearer",
		ExpiresAt: claims.ExpiresAt, IssuedAt: claims.IssuedAt, NotBefore: claims.NotBefore, Audience: claims.Audience,
		Issuer: claims.Issuer, JWTID: claims.JWTID, Email: claims.Email, EmailVerified: claims.EmailVerified,
		AuthorizationProfile: claims.AuthorizationProfile,
	}
}

func signJWTForTest(t *testing.T, key *ecdsa.PrivateKey, claims jwtAccessClaims) string {
	t.Helper()
	headerRaw, _ := json.Marshal(jwtHeader{Algorithm: jwtAccessTokenAlgorithm, KeyID: "auth-key-202607", Type: jwtAccessTokenType})
	claimsRaw, _ := json.Marshal(claims)
	signingInput := base64.RawURLEncoding.EncodeToString(headerRaw) + "." + base64.RawURLEncoding.EncodeToString(claimsRaw)
	digest := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func jwtShapedTokenForTest(t *testing.T, payload string) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES256","kid":"auth-key-202607","typ":"at+jwt"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return header + "." + claims + ".invalid-signature"
}
