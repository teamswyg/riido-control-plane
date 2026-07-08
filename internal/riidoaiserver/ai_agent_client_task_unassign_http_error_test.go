package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientUnassignTaskRejectsBoundaryErrors(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: store})
	t.Cleanup(assignmentStore.Close)
	server := NewServer(ServerConfig{
		AIAgentClient: store,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	path := "/v1/client/ai-agent/tasks/task-unassign-boundary/assignment"
	cases := []struct {
		name, token, body string
		want              int
	}{
		{"missing auth", "", `{"agent_id":"agent-public-openclaw"}`, http.StatusUnauthorized},
		{"malformed json", "ai-agent-token", `{`, http.StatusBadRequest},
		{"missing agent", "ai-agent-token", `{}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := serveUnassignTaskBoundary(server, path, tc.token, tc.body)
			if resp.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
			threads, err := store.ListAIAgentTaskThreads(t.Context(), AuthorizationResult{}, "task-unassign-boundary")
			if err != nil || len(threads.Threads) != 0 {
				t.Fatalf("rejected unassign mutated threads=%+v err=%v", threads, err)
			}
		})
	}
}

func TestHTTPAIAgentClientUnassignTaskSurfacesStoreFailure(t *testing.T) {
	errStore := unassignActionErrorStore{
		DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
		err:                           errors.New("unassign projection failed"),
	}
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: errStore})
	t.Cleanup(assignmentStore.Close)
	server := NewServer(ServerConfig{
		AIAgentClient: errStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	path := "/v1/client/ai-agent/tasks/task-unassign-boundary/assignment"
	resp := serveUnassignTaskBoundary(server, path, "ai-agent-token", `{"agent_id":"agent-public-openclaw"}`)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), errStore.err.Error()) {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
}
