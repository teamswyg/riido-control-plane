package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientCreateTaskThreadMessageSurfacesAssignFailure(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	root := seedTaskThreadMessageRoot(t, store, "task-thread-message-assign-fails")
	errStore := taskThreadMessageAssignErrorStore{
		handlerAssignmentStore: &handlerAssignmentStore{},
		err:                    errors.New("thread message queue failed"),
	}
	server := NewServer(ServerConfig{
		AIAgentClient: store,
		Assignment:    errStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	path := "/v1/client/ai-agent/tasks/" + root.TaskID + "/threads/" + root.ThreadID + "/messages"
	resp := serveTaskThreadMessageBoundary(server, path, "ai-agent-token", `{"body":"continue"}`)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), errStore.err.Error()) {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	assertTaskThreadMessageRootOnly(t, store, root)
}

func TestHTTPAIAgentClientCreateTaskThreadMessageSurfacesActionFailure(t *testing.T) {
	actionStore := taskThreadMessageActionErrorStore{
		DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
		err:                           errors.New("thread message projection failed"),
	}
	root := seedTaskThreadMessageRoot(t, actionStore, "task-thread-message-action-fails")
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: actionStore})
	t.Cleanup(assignmentStore.Close)
	server := NewServer(ServerConfig{
		AIAgentClient: actionStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	path := "/v1/client/ai-agent/tasks/" + root.TaskID + "/threads/" + root.ThreadID + "/messages"
	resp := serveTaskThreadMessageBoundary(server, path, "ai-agent-token", `{"body":"continue"}`)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), actionStore.err.Error()) {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	assertTaskThreadMessageRootOnly(t, actionStore, root)
}
