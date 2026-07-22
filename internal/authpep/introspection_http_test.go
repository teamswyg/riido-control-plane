package authpep

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHTTPIntrospectionStatusResolverBindsClientPathBodyAndProfile(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	secret := []byte("control-plane-introspection-secret-material")
	httpClient := &http.Client{Timeout: 5 * time.Second, Transport: authRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if request.URL.String() != "https://auth.riido.io/oauth/introspect" || request.Method != http.MethodPost || request.Header.Get("Content-Type") != "application/json" ||
			request.Header.Get(authConfidentialClientIDHeader) != "riido-control-plane" ||
			!validAuthClientAssertion(secret, "riido-control-plane", now, request, body) ||
			string(body) != `{"token":"header.payload.signature","token_type_hint":"access_token"}` {
			t.Fatalf("request is not identity/path/body bound: url=%s headers=%v body=%s", request.URL, request.Header, body)
		}
		return authIntrospectionResponse(http.StatusOK, `{"active":true,"scope":"ai-agent:read email openid","client_id":"riido-interaction","sub":"human:jykim@swyg.im","token_type":"Bearer","exp":1784030700,"iat":1784030400,"nbf":1784030400,"aud":"https://ai-api.riido.io","iss":"https://auth.riido.io","jti":"token-id","email":"jykim@swyg.im","email_verified":true,"authorization_profile":"riido-control-plane.production.v1"}`), nil
	})}
	resolver, err := NewHTTPIntrospectionStatusResolver(httpClient, "https://auth.riido.io", "riido-control-plane", secret, "https://ai-api.riido.io", "riido-control-plane.production.v1", func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	status, err := resolver.ResolveAccessToken(context.Background(), "header.payload.signature")
	if err != nil || !status.Active || status.Subject != "human:jykim@swyg.im" {
		t.Fatalf("status=%+v err=%v", status, err)
	}
}

func TestHTTPIntrospectionStatusResolverFailsClosedOnUntrustedResponse(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	valid := `{"active":true,"scope":"ai-agent:read email openid","client_id":"riido-interaction","sub":"human:jykim@swyg.im","token_type":"Bearer","exp":1784030700,"iat":1784030400,"nbf":1784030400,"aud":"https://ai-api.riido.io","iss":"https://auth.riido.io","jti":"token-id","email":"jykim@swyg.im","email_verified":true,"authorization_profile":"riido-control-plane.production.v1"}`
	cases := map[string]string{
		"wrong issuer":        strings.Replace(valid, "https://auth.riido.io", "https://evil.example", 1),
		"wrong audience":      strings.Replace(valid, "https://ai-api.riido.io", "https://work.riido.io", 1),
		"wrong profile":       strings.Replace(valid, "riido-control-plane.production.v1", "riido-ci.production.v1", 1),
		"wildcard scope":      strings.Replace(valid, "ai-agent:read", "ai-agent:*", 1),
		"unsorted scope":      strings.Replace(valid, "ai-agent:read email openid", "openid email ai-agent:read", 1),
		"claim leak inactive": `{"active":false,"sub":"human:jykim@swyg.im"}`,
		"unknown field":       strings.Replace(valid, `"active":true`, `"active":true,"admin":true`, 1),
		"multiple values":     valid + `{}`,
	}
	for name, responseBody := range cases {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{Timeout: 5 * time.Second, Transport: authRoundTripperFunc(func(*http.Request) (*http.Response, error) {
				return authIntrospectionResponse(http.StatusOK, responseBody), nil
			})}
			resolver, err := NewHTTPIntrospectionStatusResolver(client, "https://auth.riido.io", "riido-control-plane", []byte("control-plane-introspection-secret-material"), "https://ai-api.riido.io", "riido-control-plane.production.v1", func() time.Time { return now })
			if err != nil {
				t.Fatal(err)
			}
			if _, err := resolver.ResolveAccessToken(context.Background(), "header.payload.signature"); err == nil {
				t.Fatal("untrusted introspection response accepted")
			}
		})
	}
}

func TestHTTPIntrospectionStatusResolverRejectsRedirectAndDoesNotLeakSecrets(t *testing.T) {
	secret := "control-plane-introspection-secret-material"
	token := "sensitive.header.signature"
	calls := 0
	client := &http.Client{Timeout: 5 * time.Second, Transport: authRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		calls++
		response := authIntrospectionResponse(http.StatusTemporaryRedirect, "")
		response.Header.Set("Location", "https://untrusted.example/collect")
		return response, nil
	})}
	resolver, err := NewHTTPIntrospectionStatusResolver(client, "https://auth.riido.io", "riido-control-plane", []byte(secret), "https://ai-api.riido.io", "riido-control-plane.production.v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = resolver.ResolveAccessToken(context.Background(), token)
	if err == nil || calls != 1 {
		t.Fatalf("redirect was not refused: calls=%d err=%v", calls, err)
	}
	if strings.Contains(err.Error(), token) || strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaked credential material: %v", err)
	}
}

func TestHTTPIntrospectionStatusResolverRequiresExactBoundedConfiguration(t *testing.T) {
	secret := []byte("control-plane-introspection-secret-material")
	bounded := &http.Client{Timeout: 5 * time.Second}
	cases := []struct {
		issuer, clientID, audience, profile string
		secret                              []byte
		httpClient                          *http.Client
	}{
		{issuer: "http://auth.riido.io", clientID: "riido-control-plane", secret: secret, audience: "https://ai-api.riido.io", profile: "riido-control-plane.production.v1", httpClient: bounded},
		{issuer: "https://auth.riido.io/path", clientID: "riido-control-plane", secret: secret, audience: "https://ai-api.riido.io", profile: "riido-control-plane.production.v1", httpClient: bounded},
		{issuer: "https://auth.riido.io", clientID: "Bad Client", secret: secret, audience: "https://ai-api.riido.io", profile: "riido-control-plane.production.v1", httpClient: bounded},
		{issuer: "https://auth.riido.io", clientID: "riido-control-plane", secret: []byte("short"), audience: "https://ai-api.riido.io", profile: "riido-control-plane.production.v1", httpClient: bounded},
		{issuer: "https://auth.riido.io", clientID: "riido-control-plane", secret: secret, audience: "riido-control-plane", profile: "riido-control-plane.production.v1", httpClient: bounded},
		{issuer: "https://auth.riido.io", clientID: "riido-control-plane", secret: secret, audience: "https://ai-api.riido.io", profile: "*", httpClient: bounded},
		{issuer: "https://auth.riido.io", clientID: "riido-control-plane", secret: secret, audience: "https://ai-api.riido.io", profile: "riido-control-plane.production.v1", httpClient: &http.Client{}},
	}
	for index, value := range cases {
		if _, err := NewHTTPIntrospectionStatusResolver(value.httpClient, value.issuer, value.clientID, value.secret, value.audience, value.profile, nil); err == nil {
			t.Fatalf("invalid configuration %d accepted", index)
		}
	}
}

func validAuthClientAssertion(secret []byte, clientID string, now time.Time, request *http.Request, body []byte) bool {
	digest := sha256.Sum256(body)
	timestamp := strconv.FormatInt(now.Unix(), 10)
	message := clientID + "\n" + timestamp + "\n" + request.Method + "\n" + request.URL.Path + "\n" + hex.EncodeToString(digest[:])
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return request.Header.Get("X-Riido-Client-Timestamp") == timestamp && hmac.Equal([]byte(request.Header.Get("X-Riido-Client-Assertion")), []byte(expected))
}

func authIntrospectionResponse(status int, body string) *http.Response {
	header := http.Header{"Cache-Control": {"no-store"}}
	header.Set("Content-Type", "application/json")
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body))}
}

type authRoundTripperFunc func(*http.Request) (*http.Response, error)

func (function authRoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
