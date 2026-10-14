package authpep

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	authIntrospectionPath          = "/oauth/introspect"
	authConfidentialClientIDHeader = "X-Riido-Client-Id"
	maxAuthIntrospectionBody       = 16 << 10
)

// HTTPIntrospectionStatusResolver is the consumer-owned anti-corruption
// adapter for the Riido Auth introspection wire protocol. It deliberately
// depends on no Auth implementation package: Auth owns token state and this
// adapter owns only transport, confidential-client proof and response
// projection into the control-plane status port.
type HTTPIntrospectionStatusResolver struct {
	httpClient       *http.Client
	endpoint         string
	issuer           string
	clientID         string
	secret           []byte
	expectedAudience string
	expectedProfile  string
	clock            func() time.Time
}

type authIntrospectionRequest struct {
	Token         string `json:"token"`
	TokenTypeHint string `json:"token_type_hint"`
}

// NewHTTPIntrospectionStatusResolver binds the adapter to one exact Auth
// issuer, resource/profile pair and confidential client identity.
func NewHTTPIntrospectionStatusResolver(httpClient *http.Client, issuer, clientID string, secret []byte, expectedAudience, expectedProfile string, clock func() time.Time) (*HTTPIntrospectionStatusResolver, error) {
	if !exactHTTPSOrigin(issuer) || !exactHTTPSOrigin(expectedAudience) || !canonicalJWTIdentifier(expectedProfile, 128) || strings.Contains(expectedProfile, "*") {
		return nil, errors.New("exact Auth issuer, resource and authorization profile are required")
	}
	if !canonicalConfidentialClientID(clientID) || len(secret) < 32 || bytes.ContainsAny(secret, "\r\n\x00") {
		return nil, errors.New("canonical introspection client identity and secret are required")
	}
	if httpClient == nil || httpClient.Timeout <= 0 || httpClient.Timeout > 10*time.Second {
		return nil, errors.New("bounded introspection HTTP client is required")
	}
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	client := *httpClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return errors.New("auth introspection redirect refused")
	}
	return &HTTPIntrospectionStatusResolver{
		httpClient: &client, endpoint: issuer + authIntrospectionPath, issuer: issuer, clientID: clientID,
		secret: append([]byte(nil), secret...), expectedAudience: expectedAudience, expectedProfile: expectedProfile, clock: clock,
	}, nil
}

// ResolveAccessToken implements AccessTokenStatusResolver. Errors contain no
// access-token or client-secret material.
func (r *HTTPIntrospectionStatusResolver) ResolveAccessToken(ctx context.Context, token string) (AccessTokenStatus, error) {
	if r == nil || r.httpClient == nil {
		return AccessTokenStatus{}, errors.New("auth introspection adapter is not configured")
	}
	if token == "" || len(token) > maxAuthIntrospectionBody || strings.TrimSpace(token) != token || strings.ContainsAny(token, "\r\n\x00") {
		return AccessTokenStatus{}, errors.New("bounded access token is required")
	}
	body, err := json.Marshal(authIntrospectionRequest{Token: token, TokenTypeHint: "access_token"})
	if err != nil {
		return AccessTokenStatus{}, errors.New("encode Auth introspection request")
	}
	timestamp, assertion, err := signAuthClientRequest(r.secret, r.clientID, r.clock(), http.MethodPost, authIntrospectionPath, body)
	if err != nil {
		return AccessTokenStatus{}, fmt.Errorf("sign Auth introspection request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(body))
	if err != nil {
		return AccessTokenStatus{}, errors.New("create Auth introspection request")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set(authConfidentialClientIDHeader, r.clientID)
	request.Header.Set("X-Riido-Client-Timestamp", timestamp)
	request.Header.Set("X-Riido-Client-Assertion", assertion)
	response, err := r.httpClient.Do(request)
	if err != nil {
		return AccessTokenStatus{}, fmt.Errorf("perform Auth introspection request: %w", err)
	}
	defer response.Body.Close()
	limited, err := io.ReadAll(io.LimitReader(response.Body, maxAuthIntrospectionBody+1))
	if err != nil {
		return AccessTokenStatus{}, errors.New("read Auth introspection response")
	}
	if len(limited) > maxAuthIntrospectionBody {
		return AccessTokenStatus{}, errors.New("auth introspection response exceeded size limit")
	}
	if response.StatusCode != http.StatusOK {
		return AccessTokenStatus{}, fmt.Errorf("auth introspection returned status %d", response.StatusCode)
	}
	if media := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]); media != "application/json" || !cacheControlHasNoStore(response.Header.Values("Cache-Control")) {
		return AccessTokenStatus{}, errors.New("auth introspection returned an invalid content type or cache policy")
	}
	var status AccessTokenStatus
	decoder := json.NewDecoder(bytes.NewReader(limited))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&status); err != nil {
		return AccessTokenStatus{}, errors.New("decode Auth introspection response")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return AccessTokenStatus{}, errors.New("auth introspection returned multiple JSON values")
	}
	if !status.Active {
		if status != (AccessTokenStatus{}) {
			return AccessTokenStatus{}, errors.New("inactive Auth introspection response disclosed token claims")
		}
		return status, nil
	}
	if !r.validActiveStatus(status) {
		return AccessTokenStatus{}, errors.New("active Auth introspection response does not match the configured resource profile")
	}
	return status, nil
}

func (r *HTTPIntrospectionStatusResolver) validActiveStatus(status AccessTokenStatus) bool {
	now := r.clock().UTC().Unix()
	return status.TokenType == "Bearer" && status.Issuer == r.issuer && status.Audience == r.expectedAudience && status.AuthorizationProfile == r.expectedProfile &&
		canonicalJWTIdentifier(status.Subject, 256) && canonicalJWTIdentifier(status.ClientID, 128) && canonicalJWTIdentifier(status.JWTID, 128) && validJWTEmail(status.Email) && status.EmailVerified &&
		status.IssuedAt > 0 && status.NotBefore >= status.IssuedAt && status.ExpiresAt > status.IssuedAt && status.ExpiresAt-status.IssuedAt <= int64(15*time.Minute/time.Second) &&
		now >= status.NotBefore-30 && now < status.ExpiresAt+30 && canonicalJWTScopes(status.Scope)
}

func signAuthClientRequest(secret []byte, clientID string, timestamp time.Time, method, path string, body []byte) (string, string, error) {
	if len(secret) < 32 || !canonicalConfidentialClientID(clientID) || timestamp.IsZero() || method != http.MethodPost || path != authIntrospectionPath {
		return "", "", errors.New("complete canonical client signing input is required")
	}
	unix := strconv.FormatInt(timestamp.UTC().Unix(), 10)
	digest := sha256.Sum256(body)
	message := clientID + "\n" + unix + "\n" + method + "\n" + path + "\n" + hex.EncodeToString(digest[:])
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return unix, base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func canonicalConfidentialClientID(value string) bool {
	if value == "" || len(value) > 128 || strings.TrimSpace(value) != value {
		return false
	}
	for _, current := range value {
		if current >= 'a' && current <= 'z' || current >= '0' && current <= '9' || current == '-' || current == '_' {
			continue
		}
		return false
	}
	return true
}

// Compile-time assertion keeps the SPI relation explicit.
var _ AccessTokenStatusResolver = (*HTTPIntrospectionStatusResolver)(nil)
