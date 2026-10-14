package authpep

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	jwtAccessTokenType      = "at+jwt"
	jwtAccessTokenAlgorithm = "ES256"
	maxJWTAccessTokenBytes  = 16 << 10
)

type JWTAuthorizationConfig struct {
	Issuer               string
	Audience             string
	AuthorizationProfile string
	ClockSkew            time.Duration
	Clock                func() time.Time
}

type JWTVerificationKeyProvider interface {
	VerificationKey(context.Context, string) (*ecdsa.PublicKey, error)
}

type AccessTokenStatus struct {
	Active               bool   `json:"active"`
	Scope                string `json:"scope,omitempty"`
	ClientID             string `json:"client_id,omitempty"`
	Subject              string `json:"sub,omitempty"`
	TokenType            string `json:"token_type,omitempty"`
	ExpiresAt            int64  `json:"exp,omitempty"`
	IssuedAt             int64  `json:"iat,omitempty"`
	NotBefore            int64  `json:"nbf,omitempty"`
	Audience             string `json:"aud,omitempty"`
	Issuer               string `json:"iss,omitempty"`
	JWTID                string `json:"jti,omitempty"`
	Email                string `json:"email,omitempty"`
	EmailVerified        bool   `json:"email_verified,omitempty"`
	AuthorizationProfile string `json:"authorization_profile,omitempty"`
}

type AccessTokenStatusResolver interface {
	ResolveAccessToken(context.Context, string) (AccessTokenStatus, error)
}

// JWTAccessTokenAuthorizer is a control-plane PEP. Auth proves token
// authenticity/status; the injected domain authorizer still owns tenant and
// resource authorization and must return the same token subject.
type JWTAccessTokenAuthorizer struct {
	issuer               string
	audience             string
	authorizationProfile string
	clockSkew            time.Duration
	clock                func() time.Time
	keys                 JWTVerificationKeyProvider
	status               AccessTokenStatusResolver
	domain               RequestAuthorizer
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Type      string `json:"typ"`
}

type jwtAccessClaims struct {
	Issuer               string `json:"iss"`
	Subject              string `json:"sub"`
	Audience             string `json:"aud"`
	ExpiresAt            int64  `json:"exp"`
	IssuedAt             int64  `json:"iat"`
	NotBefore            int64  `json:"nbf"`
	JWTID                string `json:"jti"`
	ClientID             string `json:"client_id"`
	Scope                string `json:"scope"`
	Email                string `json:"email"`
	EmailVerified        bool   `json:"email_verified"`
	AuthorizationProfile string `json:"authorization_profile"`
}

func NewJWTAccessTokenAuthorizer(config JWTAuthorizationConfig, keys JWTVerificationKeyProvider, status AccessTokenStatusResolver, domain RequestAuthorizer) (*JWTAccessTokenAuthorizer, error) {
	if !exactHTTPSOrigin(config.Issuer) || !exactHTTPSOrigin(config.Audience) || !canonicalJWTIdentifier(config.AuthorizationProfile, 128) || strings.Contains(config.AuthorizationProfile, "*") {
		return nil, errors.New("exact JWT issuer, resource and authorization profile are required")
	}
	if keys == nil || status == nil || domain == nil {
		return nil, errors.New("JWT key, status and domain authorization ports are required")
	}
	if config.ClockSkew < 0 || config.ClockSkew > time.Minute {
		return nil, errors.New("JWT clock skew must be between zero and one minute")
	}
	if config.Clock == nil {
		config.Clock = func() time.Time { return time.Now().UTC() }
	}
	return &JWTAccessTokenAuthorizer{
		issuer: config.Issuer, audience: config.Audience, authorizationProfile: config.AuthorizationProfile,
		clockSkew: config.ClockSkew, clock: config.Clock, keys: keys, status: status, domain: domain,
	}, nil
}

func (a *JWTAccessTokenAuthorizer) Authorize(ctx context.Context, bearerToken string, request AuthorizationRequest) (AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthorizationResult{}, err
	}
	if a == nil {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	if !a.claimsConfiguredIssuer(bearerToken) {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	claims, err := a.verify(ctx, bearerToken)
	if err != nil {
		if errors.Is(err, ErrAuthorizationInvalidCredential) {
			return AuthorizationResult{}, err
		}
		return AuthorizationResult{}, fmt.Errorf("verify Auth access token: %w", err)
	}
	status, err := a.status.ResolveAccessToken(ctx, bearerToken)
	if err != nil {
		return AuthorizationResult{}, fmt.Errorf("resolve Auth access-token status: %w", err)
	}
	if !statusMatchesClaims(status, claims) {
		return AuthorizationResult{}, ErrAuthorizationInvalidCredential
	}
	if !jwtScopesPermit(claims.Scope, request) {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	result, err := a.domain.Authorize(ctx, bearerToken, request)
	if err != nil {
		if errors.Is(err, ErrAuthorizationForbidden) || errors.Is(err, ErrAuthorizationUnauthenticated) || errors.Is(err, ErrAuthorizationInvalidCredential) {
			return AuthorizationResult{}, ErrAuthorizationForbidden
		}
		return AuthorizationResult{}, fmt.Errorf("evaluate control-plane domain authorization: %w", err)
	}
	if strings.TrimSpace(result.PrincipalID) != claims.Subject || (request.WorkspaceID != "" && result.WorkspaceID != request.WorkspaceID) {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	return result, nil
}

func (a *JWTAccessTokenAuthorizer) verify(ctx context.Context, raw string) (jwtAccessClaims, error) {
	if raw == "" || len(raw) > maxJWTAccessTokenBytes || strings.TrimSpace(raw) != raw || strings.ContainsAny(raw, "\r\n\x00") {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	segments := strings.Split(raw, ".")
	if len(segments) != 3 || segments[0] == "" || segments[1] == "" || segments[2] == "" {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	headerRaw, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil || len(headerRaw) > 1024 {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	var header jwtHeader
	if decodeExactJWTJSON(headerRaw, &header) != nil || header.Algorithm != jwtAccessTokenAlgorithm || header.Type != jwtAccessTokenType || !canonicalJWTIdentifier(header.KeyID, 64) {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	key, err := a.keys.VerificationKey(ctx, header.KeyID)
	if err != nil {
		return jwtAccessClaims{}, err
	}
	if key == nil {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	keyBytes, keyBytesErr := key.Bytes()
	parsedKey, parseKeyErr := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), keyBytes)
	if keyBytesErr != nil || parseKeyErr != nil || len(keyBytes) != 65 || !key.Equal(parsedKey) {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	signature, err := base64.RawURLEncoding.DecodeString(segments[2])
	if err != nil || len(signature) != 64 {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	digest := sha256.Sum256([]byte(segments[0] + "." + segments[1]))
	if !ecdsa.Verify(key, digest[:], bytesToBigInt(signature[:32]), bytesToBigInt(signature[32:])) {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	claimsRaw, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil || len(claimsRaw) > 8<<10 {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	var claims jwtAccessClaims
	if decodeExactJWTJSON(claimsRaw, &claims) != nil || !a.validClaims(claims) {
		return jwtAccessClaims{}, ErrAuthorizationInvalidCredential
	}
	return claims, nil
}

func (a *JWTAccessTokenAuthorizer) validClaims(claims jwtAccessClaims) bool {
	now := a.clock().UTC()
	issued := time.Unix(claims.IssuedAt, 0).UTC()
	notBefore := time.Unix(claims.NotBefore, 0).UTC()
	expires := time.Unix(claims.ExpiresAt, 0).UTC()
	return claims.Issuer == a.issuer && claims.Audience == a.audience && claims.AuthorizationProfile == a.authorizationProfile &&
		canonicalJWTIdentifier(claims.Subject, 256) && canonicalJWTIdentifier(claims.ClientID, 128) && canonicalJWTIdentifier(claims.JWTID, 128) && validJWTEmail(claims.Email) && claims.EmailVerified &&
		canonicalJWTScopes(claims.Scope) && !notBefore.Before(issued) && expires.After(issued) && expires.Sub(issued) <= 15*time.Minute &&
		!now.Add(a.clockSkew).Before(notBefore) && now.Add(-a.clockSkew).Before(expires) && !issued.After(now.Add(a.clockSkew))
}

func statusMatchesClaims(status AccessTokenStatus, claims jwtAccessClaims) bool {
	return status.Active && status.TokenType == "Bearer" && status.Issuer == claims.Issuer && status.Subject == claims.Subject && status.Audience == claims.Audience &&
		status.ExpiresAt == claims.ExpiresAt && status.IssuedAt == claims.IssuedAt && status.NotBefore == claims.NotBefore && status.JWTID == claims.JWTID &&
		status.ClientID == claims.ClientID && status.Scope == claims.Scope && status.Email == claims.Email && status.EmailVerified == claims.EmailVerified &&
		status.AuthorizationProfile == claims.AuthorizationProfile
}

func jwtScopesPermit(scope string, request AuthorizationRequest) bool {
	required, ok := exactJWTAuthorizationScope(request)
	if !ok {
		return false
	}
	for _, candidate := range strings.Split(scope, " ") {
		if candidate == required {
			return true
		}
	}
	return false
}

func exactJWTAuthorizationScope(request AuthorizationRequest) (string, bool) {
	switch request.Resource {
	case AuthorizationResourceAIAgentClient:
		switch request.Action {
		case AuthorizationActionRead, AuthorizationActionDeviceRead, AuthorizationActionStream:
			return "ai-agent:read", true
		case AuthorizationActionCreate, AuthorizationActionAssign, AuthorizationActionStop, AuthorizationActionUpdate, AuthorizationActionDelete, AuthorizationActionDeviceControl:
			return "ai-agent:write", true
		default:
			return "", false
		}
	case AuthorizationResourceMetrics:
		if request.Action == AuthorizationActionRead {
			return "metrics:read", true
		}
	case AuthorizationResourceComponentTask:
		return "component-task:" + string(request.Action), request.Action != ""
	case AuthorizationResourceComponentTaskEvents:
		return "component-task-events:" + string(request.Action), request.Action != ""
	case AuthorizationResourceAgent:
		return "agent:" + string(request.Action), request.Action != ""
	case AuthorizationResourceAgentCatalog:
		return "agent-catalog:" + string(request.Action), request.Action != ""
	}
	return "", false
}

func canonicalJWTScopes(value string) bool {
	values := strings.Fields(value)
	if len(values) < 2 || len(values) > 32 || strings.Join(values, " ") != value || !sort.StringsAreSorted(values) {
		return false
	}
	identity := map[string]bool{}
	for index, current := range values {
		if !canonicalJWTIdentifier(current, 128) || strings.Contains(current, "*") || (index > 0 && values[index-1] == current) {
			return false
		}
		if current == "openid" || current == "email" {
			identity[current] = true
		}
	}
	return identity["openid"] && identity["email"]
}

func canonicalJWTIdentifier(value string, maximum int) bool {
	if value == "" || len(value) > maximum || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\r\n\x00 ") {
		return false
	}
	for _, current := range value {
		if current >= 'a' && current <= 'z' || current >= 'A' && current <= 'Z' || current >= '0' && current <= '9' || strings.ContainsRune("-._:@/", current) {
			continue
		}
		return false
	}
	return true
}

func exactHTTPSOrigin(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.Host == parsed.Hostname() && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == "" && parsed.User == nil && parsed.String() == value
}

func bytesToBigInt(value []byte) *big.Int { return new(big.Int).SetBytes(value) }
