package authpep

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxJWKSResponseBytes = 64 << 10

// HTTPJWKSProvider consumes only the Auth issuer's exact ES256/P-256 key set.
// Stale keys are never used after max-age and an unknown kid forces one
// conditional refresh so rotation can converge without widening the allowlist.
type HTTPJWKSProvider struct {
	endpoint   string
	httpClient *http.Client
	clock      func() time.Time

	mu        sync.Mutex
	keys      map[string]*ecdsa.PublicKey
	etag      string
	expiresAt time.Time
}

type jwksDocument struct {
	Keys []jwkVerificationKey `json:"keys"`
}

type jwkVerificationKey struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Curve     string `json:"crv"`
	X         string `json:"x"`
	Y         string `json:"y"`
}

func NewHTTPJWKSProvider(issuer string, httpClient *http.Client, clock func() time.Time) (*HTTPJWKSProvider, error) {
	if !exactHTTPSOrigin(issuer) {
		return nil, errors.New("JWKS issuer must be an exact HTTPS origin")
	}
	if httpClient == nil || httpClient.Timeout <= 0 || httpClient.Timeout > 10*time.Second {
		return nil, errors.New("bounded JWKS HTTP client is required")
	}
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &HTTPJWKSProvider{endpoint: issuer + "/.well-known/jwks.json", httpClient: httpClient, clock: clock, keys: map[string]*ecdsa.PublicKey{}}, nil
}

func (p *HTTPJWKSProvider) VerificationKey(ctx context.Context, keyID string) (*ecdsa.PublicKey, error) {
	if p == nil || !canonicalJWTIdentifier(keyID, 64) {
		return nil, ErrAuthorizationInvalidCredential
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	now := p.clock().UTC()
	if now.Before(p.expiresAt) {
		if key, ok := p.keys[keyID]; ok {
			return cloneECDSAKey(key)
		}
		// A valid but unknown kid may be the newly active epoch. Force one
		// conditional refresh even while the old cache is fresh.
		if err := p.refresh(ctx, now); err != nil {
			return nil, err
		}
	} else if err := p.refresh(ctx, now); err != nil {
		return nil, err
	}
	key, ok := p.keys[keyID]
	if !ok {
		return nil, ErrAuthorizationInvalidCredential
	}
	return cloneECDSAKey(key)
}

func (p *HTTPJWKSProvider) refresh(ctx context.Context, now time.Time) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/jwk-set+json")
	if p.etag != "" {
		request.Header.Set("If-None-Match", p.etag)
	}
	response, err := p.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("fetch Auth JWKS: %w", err)
	}
	defer response.Body.Close()
	maxAge, err := boundedJWKSMaxAge(response.Header.Get("Cache-Control"))
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotModified {
		if p.etag == "" || len(p.keys) == 0 {
			return errors.New("auth JWKS returned 304 without a cached key set")
		}
		p.expiresAt = now.Add(maxAge)
		return nil
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("auth JWKS returned status %d", response.StatusCode)
	}
	if media := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]); media != "application/jwk-set+json" {
		return errors.New("auth JWKS returned an invalid content type")
	}
	etag := strings.TrimSpace(response.Header.Get("ETag"))
	if len(etag) < 3 || !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) || strings.ContainsAny(etag, "\r\n\x00") {
		return errors.New("auth JWKS requires a strong ETag")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxJWKSResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read Auth JWKS: %w", err)
	}
	if len(body) > maxJWKSResponseBytes {
		return errors.New("auth JWKS exceeded size limit")
	}
	keys, err := decodeJWKS(body)
	if err != nil {
		return err
	}
	p.keys = keys
	p.etag = etag
	p.expiresAt = now.Add(maxAge)
	return nil
}

func decodeJWKS(body []byte) (map[string]*ecdsa.PublicKey, error) {
	var document jwksDocument
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return nil, errors.New("auth JWKS document is invalid")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF || len(document.Keys) == 0 || len(document.Keys) > 3 {
		return nil, errors.New("auth JWKS key set is invalid")
	}
	keys := make(map[string]*ecdsa.PublicKey, len(document.Keys))
	for _, candidate := range document.Keys {
		if candidate.KeyType != "EC" || candidate.Use != "sig" || candidate.Algorithm != jwtAccessTokenAlgorithm || candidate.Curve != "P-256" || !canonicalJWTIdentifier(candidate.KeyID, 64) {
			return nil, errors.New("auth JWKS contains an unsupported key")
		}
		if _, duplicate := keys[candidate.KeyID]; duplicate {
			return nil, errors.New("auth JWKS contains a duplicate kid")
		}
		x, errX := base64.RawURLEncoding.DecodeString(candidate.X)
		y, errY := base64.RawURLEncoding.DecodeString(candidate.Y)
		if errX != nil || errY != nil || len(x) != 32 || len(y) != 32 {
			return nil, errors.New("auth JWKS contains invalid P-256 coordinates")
		}
		point := make([]byte, 65)
		point[0] = 4
		copy(point[1:33], x)
		copy(point[33:], y)
		key, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), point)
		if err != nil {
			return nil, errors.New("auth JWKS public key is not on P-256")
		}
		keys[candidate.KeyID] = key
	}
	return keys, nil
}

func boundedJWKSMaxAge(value string) (time.Duration, error) {
	var found bool
	var seconds int64
	for _, directive := range strings.Split(value, ",") {
		parts := strings.SplitN(strings.TrimSpace(directive), "=", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "max-age") {
			if found {
				return 0, errors.New("auth JWKS cache-control repeats max-age")
			}
			parsed, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || parsed < 1 || parsed > 60 {
				return 0, errors.New("auth JWKS max-age must be between 1 and 60 seconds")
			}
			found, seconds = true, parsed
		}
	}
	if !found {
		return 0, errors.New("auth JWKS cache-control requires max-age")
	}
	return time.Duration(seconds) * time.Second, nil
}

func cloneECDSAKey(key *ecdsa.PublicKey) (*ecdsa.PublicKey, error) {
	encoded, err := key.Bytes()
	if err != nil {
		return nil, errors.New("auth JWKS cached key is invalid")
	}
	clone, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), encoded)
	if err != nil || !key.Equal(clone) {
		return nil, errors.New("auth JWKS cached key is invalid")
	}
	return clone, nil
}
