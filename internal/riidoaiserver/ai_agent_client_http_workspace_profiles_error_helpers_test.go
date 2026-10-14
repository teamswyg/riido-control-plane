package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type workspaceProfilesErrorStore struct {
	*DevelopmentAIAgentClientStore
	err          error
	reconcileErr error
}

func (s workspaceProfilesErrorStore) ListWorkspaceAssignedAgentProfiles(
	context.Context,
	AuthorizationResult,
) (AssignedAgentProfileMapResponse, error) {
	return AssignedAgentProfileMapResponse{}, s.err
}

func (s workspaceProfilesErrorStore) ReconcileAIAgentActiveThreadProjections(
	context.Context,
	AuthorizationResult,
	string,
	AssignmentProjectionReader,
) (bool, error) {
	return false, s.reconcileErr
}

func newWorkspaceProfilesErrorTestServer(
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
