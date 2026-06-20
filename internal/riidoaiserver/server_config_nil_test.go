package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServerTreatsTypedNilAssignmentStoreAsUnconfigured(t *testing.T) {
	var store *Store
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "metrics-reader",
		Token:       "metrics-token",
		Scopes:      []string{"metrics:read"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(ServerConfig{Assignment: store, Authorizer: auth}).Handler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer metrics-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	if strings.Contains(resp.Body.String(), "panic") {
		t.Fatalf("typed nil assignment leaked panic wording: %s", resp.Body.String())
	}
}
