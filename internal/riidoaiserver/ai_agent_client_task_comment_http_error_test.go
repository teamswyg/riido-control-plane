package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type taskCommentErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s taskCommentErrorStore) SubmitAIAgentTaskComment(context.Context, AuthorizationResult, string, SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error) {
	return AIAgentTaskActionResponse{}, s.err
}

func TestHTTPAIAgentClientSubmitTaskCommentRejectsBoundaryErrors(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	server := newTaskThreadReadErrorTestServer(t, nil, store, nil)
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-comment/comments"
	cases := []struct {
		name, token, body string
		want              int
	}{
		{"missing auth", "", `{"agent_id":"agent-owned-codex","body":"continue"}`, http.StatusUnauthorized},
		{"malformed json", "user-token", `{`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := serveAssignTaskBoundary(server, base, tc.token, tc.body)
			if resp.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
			threads, err := store.ListAIAgentTaskThreads(t.Context(), AuthorizationResult{WorkspaceID: defaultAIAgentClientWorkspaceID}, "task-comment")
			if err != nil || len(threads.Threads) != 0 {
				t.Fatalf("rejected comment mutated threads=%+v err=%v", threads, err)
			}
		})
	}
}

func TestHTTPAIAgentClientSubmitTaskCommentSurfacesStoreFailures(t *testing.T) {
	errStore := taskCommentErrorStore{
		DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
		err:                           errors.New("comment projection failed"),
	}
	server := newTaskThreadReadErrorTestServer(t, nil, errStore, nil)
	path := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-comment/comments"
	resp := serveAssignTaskBoundary(server, path, "user-token", `{"agent_id":"agent-owned-codex","body":"continue"}`)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), errStore.err.Error()) {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPAIAgentClientSubmitTaskCommentMapsMissingAgentToNotFound(t *testing.T) {
	errStore := taskCommentErrorStore{
		DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
		err:                           ErrAIAgentNotFound,
	}
	server := newTaskThreadReadErrorTestServer(t, nil, errStore, nil)
	path := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-comment/comments"
	resp := serveAssignTaskBoundary(server, path, "user-token", `{"agent_id":"agent-missing","body":"continue"}`)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("status=%d want=%d body=%s", resp.Code, http.StatusNotFound, resp.Body.String())
	}
}
