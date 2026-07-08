package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientCreateTaskThreadMessageRejectsBoundaryErrors(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: store})
	t.Cleanup(assignmentStore.Close)
	root := seedTaskThreadMessageRoot(t, store, "task-thread-message-boundary")
	server := NewServer(ServerConfig{
		AIAgentClient: store,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	path := "/v1/client/ai-agent/tasks/" + root.TaskID + "/threads/" + root.ThreadID + "/messages"
	cases := []struct {
		name, token, body string
		want              int
	}{
		{"missing auth", "", `{"body":"continue"}`, http.StatusUnauthorized},
		{"malformed json", "ai-agent-token", `{`, http.StatusBadRequest},
		{"missing body", "ai-agent-token", `{}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := serveTaskThreadMessageBoundary(server, path, tc.token, tc.body)
			if resp.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
			assertTaskThreadMessageRootOnly(t, store, root)
		})
	}
}

func TestHTTPAIAgentClientCreateTaskThreadMessageRequiresAssignmentStore(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	root := seedTaskThreadMessageRoot(t, store, "task-thread-message-no-assignment")
	server := NewServer(ServerConfig{
		AIAgentClient: store,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	path := "/v1/client/ai-agent/tasks/" + root.TaskID + "/threads/" + root.ThreadID + "/messages"
	resp := serveTaskThreadMessageBoundary(server, path, "ai-agent-token", `{"body":"continue"}`)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=%d body=%s", resp.Code, http.StatusServiceUnavailable, resp.Body.String())
	}
	assertTaskThreadMessageRootOnly(t, store, root)
}
