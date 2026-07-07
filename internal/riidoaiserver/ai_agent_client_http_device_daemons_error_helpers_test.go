package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type failingDeviceDaemonsStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s *failingDeviceDaemonsStore) ListAIAgentDeviceDaemons(
	context.Context,
	AuthorizationResult,
	string,
) (DeviceDaemonListResponse, error) {
	return DeviceDaemonListResponse{}, s.err
}

func newDeviceDaemonsErrorTestServer(t *testing.T, scopes []string, store AIAgentClientStore) http.Handler {
	t.Helper()
	if len(scopes) == 0 {
		scopes = []string{"ai-agent:*"}
	}
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      scopes,
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return NewServer(ServerConfig{AIAgentClient: store, Authorizer: auth}).Handler()
}
