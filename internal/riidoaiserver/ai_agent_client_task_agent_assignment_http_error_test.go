package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type createTaskAssignmentActionErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s createTaskAssignmentActionErrorStore) CreateAIAgentTaskAgentAssignment(context.Context, AuthorizationResult, string, AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	return AIAgentTaskActionResponse{}, s.err
}

func TestHTTPAIAgentClientCreateTaskAgentAssignmentRejectsBoundaryErrors(t *testing.T) {
	server, store := newAssignTaskBoundaryServer(t, true)
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-boundary/agent-assignments"
	cases := []struct {
		name, path, token, body string
		want                    int
	}{
		{"unscoped route", "/v1/client/ai-agent/tasks/task-boundary/agent-assignments", "user-token", `{"agent_id":"agent-owned-codex"}`, http.StatusNotFound},
		{"missing auth", base, "", `{"agent_id":"agent-owned-codex"}`, http.StatusUnauthorized},
		{"malformed json", base, "user-token", `{`, http.StatusBadRequest},
		{"missing agent", base, "user-token", `{}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := serveAssignTaskBoundary(server, tc.path, tc.token, tc.body)
			if resp.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
			assertAssignTaskBoundaryPollNone(t, store)
		})
	}
}

func TestHTTPAIAgentClientCreateTaskAgentAssignmentRequiresAssignmentStore(t *testing.T) {
	server, _ := newAssignTaskBoundaryServer(t, false)
	path := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-boundary/agent-assignments"
	resp := serveAssignTaskBoundary(server, path, "user-token", `{"agent_id":"agent-owned-codex"}`)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=%d body=%s", resp.Code, http.StatusServiceUnavailable, resp.Body.String())
	}
}

func TestHTTPAIAgentClientCreateTaskAgentAssignmentSurfacesStoreFailures(t *testing.T) {
	actionStore := createTaskAssignmentActionErrorStore{
		DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
		err:                           errors.New("projection write failed"),
	}
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: actionStore})
	t.Cleanup(func() { assignmentStore.Close() })
	server := NewServer(ServerConfig{
		AIAgentClient: actionStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	path := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-boundary/agent-assignments"
	resp := serveAssignTaskBoundary(server, path, "ai-agent-token", `{"agent_id":"agent-owned-codex"}`)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), actionStore.err.Error()) {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
}
