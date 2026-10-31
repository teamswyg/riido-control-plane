package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

// The legacy PDP verifies the app JWT signature and workspace membership, but
// has no CLI grant/installation lookup. A signed CLI token must not reach it.
func TestAuthorizerFromEnvRejectsMonolithCLIBeforeLegacyPDP(t *testing.T) {
	for _, profile := range []string{"legacy-only", "dual-auth"} {
		t.Run(profile, func(t *testing.T) {
			clearRiidoAIServerEnv(t)
			var calls atomic.Int32
			secret := []byte("synthetic-local-legacy-secret")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				var body struct {
					BearerToken string `json:"bearer_token"`
				}
				if json.NewDecoder(r.Body).Decode(&body) != nil {
					t.Error("invalid PDP request")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				parts := strings.Split(body.BearerToken, ".")
				if len(parts) != 3 {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				mac := hmac.New(sha256.New, secret)
				_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
				sig, err := base64.RawURLEncoding.DecodeString(parts[2])
				if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"schema_version": riidoaiserver.ExternalAuthorizerResponseSchemaVersion, "allowed": true, "principal_id": "synthetic-user"})
			}))
			defer server.Close()
			t.Setenv(envExternalAuthzURL, server.URL)
			if profile == "dual-auth" {
				setAuthProfileEnv(t)
			}
			authorizer, err := authorizerFromEnv()
			if err != nil {
				t.Fatal(err)
			}
			claims := `{"id":"synthetic-user","clientId":"riido-cli","cliGrantId":"synthetic-grant","scopes":["read"],"aud":"https://api.example.test","iss":"https://api.example.test","exp":4102444800}`
			raw := signedCLIBoundaryFixture(claims, secret)
			for _, action := range []riidoaiserver.AuthorizationAction{riidoaiserver.AuthorizationActionRead, riidoaiserver.AuthorizationActionCreate} {
				_, err := authorizer.Authorize(context.Background(), raw, riidoaiserver.AuthorizationRequest{Resource: riidoaiserver.AuthorizationResourceAIAgentClient, Action: action, WorkspaceID: "synthetic-workspace", TaskID: "synthetic-task"})
				if !errors.Is(err, riidoaiserver.ErrAuthorizationForbidden) {
					t.Errorf("CLI %s reached legacy permission: %v", action, err)
				}
			}
			if calls.Load() != 0 {
				t.Errorf("CLI made %d legacy PDP calls", calls.Load())
			}
			// A normal application token retains the existing external authorization.
			_, err = authorizer.Authorize(context.Background(), signedCLIBoundaryFixture(`{"id":"synthetic-user","exp":4102444800}`, secret), riidoaiserver.AuthorizationRequest{Resource: riidoaiserver.AuthorizationResourceAIAgentClient, Action: riidoaiserver.AuthorizationActionRead, WorkspaceID: "synthetic-workspace", TaskID: "synthetic-task"})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func signedCLIBoundaryFixture(claims string, secret []byte) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	unsigned := header + "." + payload
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
