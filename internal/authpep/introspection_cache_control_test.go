package authpep

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHTTPIntrospectionStatusResolverRequiresNoStoreResponse(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	valid := `{"active":true,"scope":"ai-agent:read email openid","client_id":"riido-interaction","sub":"human:jykim@swyg.im","token_type":"Bearer","exp":1784030700,"iat":1784030400,"nbf":1784030400,"aud":"https://ai-api.riido.io","iss":"https://auth.riido.io","jti":"token-id","email":"jykim@swyg.im","email_verified":true,"authorization_profile":"riido-control-plane.production.v1"}`
	cases := map[string][]string{
		"missing":   nil,
		"cacheable": {"private, max-age=30"},
		"lookalike": {"private, x-no-store"},
	}
	for name, cacheControl := range cases {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{Timeout: 5 * time.Second, Transport: authRoundTripperFunc(func(*http.Request) (*http.Response, error) {
				response := authIntrospectionResponse(http.StatusOK, valid)
				response.Header.Del("Cache-Control")
				for _, value := range cacheControl {
					response.Header.Add("Cache-Control", value)
				}
				return response, nil
			})}
			resolver, err := NewHTTPIntrospectionStatusResolver(client, "https://auth.riido.io", "riido-control-plane", []byte("control-plane-introspection-secret-material"), "https://ai-api.riido.io", "riido-control-plane.production.v1", func() time.Time { return now })
			if err != nil {
				t.Fatal(err)
			}
			if _, err := resolver.ResolveAccessToken(context.Background(), "header.payload.signature"); err == nil {
				t.Fatal("cacheable introspection response accepted")
			}
		})
	}
}
