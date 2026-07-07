package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientLegacyThreadMessageLookupErrors(t *testing.T) {
	server := newTaskThreadReadErrorTestServer(
		t,
		nil,
		legacyThreadMessageStore{
			DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
			findErr:                       ErrAIAgentNotFound,
		},
		nil,
	)
	resp := postLegacyThreadMessage(t, server, legacyThreadMessageWorkspacePath("missing-thread"))
	if resp.Code != http.StatusNotFound {
		t.Fatalf("status=%d want=%d body=%s", resp.Code, http.StatusNotFound, resp.Body.String())
	}
}

func TestHTTPAIAgentClientLegacyThreadMessageDelegatesAfterLookup(t *testing.T) {
	server := newTaskThreadReadErrorTestServer(
		t,
		nil,
		legacyThreadMessageStore{
			DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
			thread: AIAgentTaskThreadRecord{
				TaskID:   "task-thread-compat",
				ThreadID: "thread-compat",
			},
		},
		nil,
	)
	resp := postLegacyThreadMessage(t, server, legacyThreadMessageWorkspacePath("thread-compat"))
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=%d body=%s", resp.Code, http.StatusServiceUnavailable, resp.Body.String())
	}
}

func legacyThreadMessageWorkspacePath(threadID string) string {
	return "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/threads/" + threadID + "/messages"
}
