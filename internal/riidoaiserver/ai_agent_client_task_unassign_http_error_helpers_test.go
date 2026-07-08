package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
)

type unassignActionErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s unassignActionErrorStore) UnassignAIAgentTask(
	context.Context,
	AuthorizationResult,
	string,
	UnassignAIAgentTaskRequest,
) (AIAgentTaskActionResponse, error) {
	return AIAgentTaskActionResponse{}, s.err
}

func serveUnassignTaskBoundary(server http.Handler, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set(aiAgentTokenHeader, token)
	}
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	return resp
}
