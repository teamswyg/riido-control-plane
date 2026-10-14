package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type deleteActiveThreadsErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s deleteActiveThreadsErrorStore) ActiveAIAgentTaskThreadsForAgent(
	context.Context,
	AuthorizationResult,
	string,
) ([]AIAgentTaskThreadRecord, error) {
	return nil, s.err
}

type deleteActiveThreadsStore struct {
	*DevelopmentAIAgentClientStore
}

func (s deleteActiveThreadsStore) ActiveAIAgentTaskThreadsForAgent(
	context.Context,
	AuthorizationResult,
	string,
) ([]AIAgentTaskThreadRecord, error) {
	return []AIAgentTaskThreadRecord{{
		TaskID:       "task-delete-error",
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-delete-error",
	}}, nil
}

type deleteCancelErrorStore struct {
	*Store
	err error
}

func (s deleteCancelErrorStore) CancelAssignment(
	context.Context,
	string,
	CancelAssignmentRequest,
) (Assignment, error) {
	return Assignment{}, s.err
}

type deleteAgentErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s deleteAgentErrorStore) DeleteAIAgent(
	context.Context,
	AuthorizationResult,
	string,
) (DeleteAgentResponse, error) {
	return DeleteAgentResponse{}, s.err
}

func newDeleteErrorTestServer(t *testing.T, store AIAgentClientStore, assignment AssignmentStore) http.Handler {
	t.Helper()
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return NewServer(ServerConfig{AIAgentClient: store, Assignment: assignment, Authorizer: auth}).Handler()
}
