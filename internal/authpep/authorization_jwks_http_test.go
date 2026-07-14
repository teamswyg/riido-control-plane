package authpep

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPJWKSProviderCachesByETagAndRefreshesUnknownKID(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	first, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	second, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	calls := 0
	httpClient := &http.Client{Timeout: 5 * time.Second, Transport: jwksRoundTripper(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.URL.String() != "https://auth.riido.io/.well-known/jwks.json" || request.Header.Get("Accept") != "application/jwk-set+json" {
			t.Fatalf("unexpected request: %s headers=%v", request.URL, request.Header)
		}
		if calls == 1 {
			if request.Header.Get("If-None-Match") != "" {
				t.Fatal("first request sent an ETag")
			}
			return jwksResponse(t, http.StatusOK, `"epoch-a"`, []namedJWK{{"auth-key-a", &first.PublicKey}}), nil
		}
		if request.Header.Get("If-None-Match") != `"epoch-a"` {
			t.Fatalf("refresh If-None-Match=%q", request.Header.Get("If-None-Match"))
		}
		return jwksResponse(t, http.StatusOK, `"epoch-b"`, []namedJWK{{"auth-key-a", &first.PublicKey}, {"auth-key-b", &second.PublicKey}}), nil
	})}
	provider, err := NewHTTPJWKSProvider("https://auth.riido.io", httpClient, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.VerificationKey(t.Context(), "auth-key-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.VerificationKey(t.Context(), "auth-key-a"); err != nil || calls != 1 {
		t.Fatalf("cached key err=%v calls=%d", err, calls)
	}
	got, err := provider.VerificationKey(t.Context(), "auth-key-b")
	if err != nil || !got.Equal(&second.PublicKey) || calls != 2 {
		t.Fatalf("rotated key=%v err=%v calls=%d", got, err, calls)
	}
}

func TestHTTPJWKSProviderUses304OnlyForExistingFreshenedCache(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	calls := 0
	httpClient := &http.Client{Timeout: 5 * time.Second, Transport: jwksRoundTripper(func(*http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return jwksResponse(t, http.StatusOK, `"epoch-a"`, []namedJWK{{"auth-key-a", &key.PublicKey}}), nil
		}
		header := make(http.Header)
		header.Set("Cache-Control", "public, max-age=60")
		return &http.Response{StatusCode: http.StatusNotModified, Header: header, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	provider, _ := NewHTTPJWKSProvider("https://auth.riido.io", httpClient, func() time.Time { return now })
	if _, err := provider.VerificationKey(t.Context(), "auth-key-a"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(61 * time.Second)
	if _, err := provider.VerificationKey(t.Context(), "auth-key-a"); err != nil || calls != 2 {
		t.Fatalf("304 refresh err=%v calls=%d", err, calls)
	}
}

func TestHTTPJWKSProviderRejectsAlgorithmCoordinateAndCacheDrift(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	valid := jwksResponse(t, http.StatusOK, `"epoch-a"`, []namedJWK{{"auth-key-a", &key.PublicKey}})
	body, _ := io.ReadAll(valid.Body)
	_ = valid.Body.Close()
	keyX, _ := publicKeyCoordinates(t, &key.PublicKey)
	cases := map[string]func(*http.Response){
		"wrong algorithm": func(response *http.Response) {
			response.Body = io.NopCloser(strings.NewReader(strings.Replace(string(body), `"ES256"`, `"RS256"`, 1)))
		},
		"short coordinate": func(response *http.Response) {
			response.Body = io.NopCloser(strings.NewReader(strings.Replace(string(body), base64.RawURLEncoding.EncodeToString(keyX), "AA", 1)))
		},
		"long cache":         func(response *http.Response) { response.Header.Set("Cache-Control", "public, max-age=300") },
		"weak etag":          func(response *http.Response) { response.Header.Set("ETag", `W/"epoch-a"`) },
		"wrong content type": func(response *http.Response) { response.Header.Set("Content-Type", "application/json") },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			response := jwksResponse(t, http.StatusOK, `"epoch-a"`, []namedJWK{{"auth-key-a", &key.PublicKey}})
			t.Cleanup(func() { _ = response.Body.Close() })
			mutate(response)
			provider, _ := NewHTTPJWKSProvider("https://auth.riido.io", &http.Client{Timeout: 5 * time.Second, Transport: jwksRoundTripper(func(*http.Request) (*http.Response, error) { return response, nil })}, nil)
			if _, err := provider.VerificationKey(t.Context(), "auth-key-a"); err == nil {
				t.Fatal("untrusted JWKS response accepted")
			}
		})
	}
}

type jwksRoundTripper func(*http.Request) (*http.Response, error)

func (function jwksRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

type namedJWK struct {
	id  string
	key *ecdsa.PublicKey
}

func jwksResponse(t *testing.T, status int, etag string, keys []namedJWK) *http.Response {
	t.Helper()
	document := jwksDocument{Keys: make([]jwkVerificationKey, 0, len(keys))}
	for _, value := range keys {
		x, y := publicKeyCoordinates(t, value.key)
		document.Keys = append(document.Keys, jwkVerificationKey{
			KeyType: "EC", Use: "sig", Algorithm: "ES256", KeyID: value.id, Curve: "P-256",
			X: base64.RawURLEncoding.EncodeToString(x),
			Y: base64.RawURLEncoding.EncodeToString(y),
		})
	}
	body, _ := json.Marshal(document)
	header := make(http.Header)
	header.Set("Content-Type", "application/jwk-set+json")
	header.Set("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	header.Set("ETag", etag)
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(string(body)))}
}

func publicKeyCoordinates(t *testing.T, key *ecdsa.PublicKey) ([]byte, []byte) {
	t.Helper()
	encoded, err := key.Bytes()
	if err != nil || len(encoded) != 65 || encoded[0] != 4 {
		t.Fatalf("encode P-256 public key: len=%d err=%v", len(encoded), err)
	}
	return append([]byte(nil), encoded[1:33]...), append([]byte(nil), encoded[33:]...)
}
