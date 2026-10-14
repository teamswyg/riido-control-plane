package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type taskThreadMessageAssignErrorStore struct {
	*handlerAssignmentStore
	err error
}

func (s taskThreadMessageAssignErrorStore) AssignTask(context.Context, string, AssignRequest) (Assignment, error) {
	return Assignment{}, s.err
}

type taskThreadMessageActionErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s taskThreadMessageActionErrorStore) CreateAIAgentTaskThreadMessage(
	context.Context,
	AuthorizationResult,
	string,
	string,
	CreateAIAgentTaskThreadMessageRequest,
) (AIAgentTaskActionResponse, error) {
	return AIAgentTaskActionResponse{}, s.err
}

func seedTaskThreadMessageRoot(t *testing.T, store AIAgentClientStore, taskID string) AIAgentTaskActionResponse {
	t.Helper()
	out, err := store.AssignAIAgentTask(t.Context(), AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}, taskID, AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-root-" + taskID,
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	return out
}

func serveTaskThreadMessageBoundary(
	server http.Handler,
	path string,
	token string,
	body string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set(aiAgentTokenHeader, token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	return resp
}
