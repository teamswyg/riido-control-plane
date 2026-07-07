package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newAssignTaskBoundaryServer(t *testing.T, withAssignment bool) (http.Handler, *Store) {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "agent:agent-owned-codex:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	var assignment *Store
	if withAssignment {
		assignment = NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
		t.Cleanup(func() { assignment.Close() })
	}
	return NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignment,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler(), assignment
}

func serveAssignTaskBoundary(server http.Handler, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set(aiAgentTokenHeader, token)
	}
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	return resp
}

func assertAssignTaskBoundaryPollNone(t *testing.T, store *Store) {
	t.Helper()
	poll, err := store.PollAgent(t.Context(), "agent-owned-codex", PollRequest{
		DaemonID:  "daemon-dev-macbook",
		DeviceID:  "device-dev-macbook",
		RuntimeID: "runtime-codex-dev",
	})
	if err != nil || poll.Action != PollNone || poll.Assignment != nil {
		t.Fatalf("poll after rejected assign = %+v err=%v", poll, err)
	}
}
