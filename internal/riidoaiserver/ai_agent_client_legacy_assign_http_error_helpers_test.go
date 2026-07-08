package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type legacyAssignReplacementErrorStore struct {
	*handlerAssignmentStore
	err error
}

func (s legacyAssignReplacementErrorStore) AssignTaskReplacement(context.Context, string, AssignRequest) (Assignment, error) {
	return Assignment{}, s.err
}

type legacyAssignActionErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s legacyAssignActionErrorStore) AssignAIAgentTask(context.Context, AuthorizationResult, string, AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	return AIAgentTaskActionResponse{}, s.err
}

func legacyAssignErrorServer(t *testing.T, aiAgent AIAgentClientStore, assignment AssignmentStore) http.Handler {
	t.Helper()
	return NewServer(ServerConfig{
		AIAgentClient: aiAgent,
		Assignment:    assignment,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
}
