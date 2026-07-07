package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type clientEventsSubscriberErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s clientEventsSubscriberErrorStore) SubscribeAIAgentClientEvents(
	context.Context,
	AuthorizationResult,
) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error) {
	return nil, nil, nil, s.err
}

type clientEventsReaderOnlyErrorStore struct {
	AIAgentClientStore
	err error
}

func (s clientEventsReaderOnlyErrorStore) AIAgentClientEvents(
	context.Context,
	AuthorizationResult,
) ([]ClientStreamEvent, error) {
	return nil, s.err
}

func newClientEventsErrorTestServer(
	t *testing.T,
	scopes []string,
	store AIAgentClientStore,
) http.Handler {
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
	return NewServer(ServerConfig{
		AIAgentClient: store,
		Authorizer:    auth,
	}).Handler()
}

func clientEventsErrorTestPath() string {
	return "/v2/client/workspaces/workspace-dev-riid/ai-agent/events?replay=1"
}
