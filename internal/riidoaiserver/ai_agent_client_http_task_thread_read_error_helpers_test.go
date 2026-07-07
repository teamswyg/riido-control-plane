package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type taskThreadReadErrorStore struct {
	*DevelopmentAIAgentClientStore
	listErr         error
	subscriptionErr error
	reconcileErr    error
}

func (s taskThreadReadErrorStore) ListAIAgentTaskThreads(
	context.Context,
	AuthorizationResult,
	string,
) (AIAgentTaskThreadCollectionResponse, error) {
	return AIAgentTaskThreadCollectionResponse{}, s.listErr
}

func (s taskThreadReadErrorStore) GetAIAgentTaskThreadStreamSubscription(
	context.Context,
	AuthorizationResult,
	string,
) (AIAgentTaskThreadStreamSubscriptionResponse, error) {
	return AIAgentTaskThreadStreamSubscriptionResponse{}, s.subscriptionErr
}

func (s taskThreadReadErrorStore) ReconcileAIAgentActiveThreadProjections(
	context.Context,
	AuthorizationResult,
	string,
	AssignmentProjectionReader,
) (bool, error) {
	return false, s.reconcileErr
}

func newTaskThreadReadErrorTestServer(
	t *testing.T,
	scopes []string,
	store AIAgentClientStore,
	assignment AssignmentStore,
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
		Assignment:    assignment,
		Authorizer:    auth,
	}).Handler()
}
