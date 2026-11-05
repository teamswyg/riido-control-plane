package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type reactionTestAuthorizer struct{ principal AuthorizationResult }

func (a reactionTestAuthorizer) Authorize(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
	return a.principal, nil
}
func TestMessageReactionHTTPAuthorizesAndPersistsActor(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root, err := store.AssignAIAgentTask(t.Context(), principal, "reaction-http", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"})
	if err != nil {
		t.Fatal(err)
	}
	store.taskThreadMessages[root.ThreadID] = []AIAgentTaskThreadHistoryMessage{{MessageID: "stable-message", Role: AIAgentTaskThreadMessageRoleAgent, Body: "answer"}}
	server := NewServer(ServerConfig{AIAgentClient: store, Authorizer: reactionTestAuthorizer{principal}})
	path := "/v2/client/workspaces/" + principal.WorkspaceID + "/ai-agent/tasks/" + root.TaskID + "/threads/" + root.ThreadID + "/messages/stable-message/reactions"
	request := func(method, path, body string, authorized bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if authorized {
			r.Header.Set("Authorization", "Bearer test")
		}
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		return w
	}
	if got := request(http.MethodPost, path, `{"emoji":"2705"}`, false); got.Code != 401 {
		t.Fatalf("unauthorized=%d", got.Code)
	}
	got := request(http.MethodPost, path, `{"emoji":"2705"}`, true)
	if got.Code != 200 {
		t.Fatalf("post=%d %s", got.Code, got.Body)
	}
	var response AIAgentTaskThreadMessageReactionsResponse
	if err := json.Unmarshal(got.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Reactions) != 1 || response.Reactions[0].Actors[0] != (AIAgentMessageReactionActor{Kind: "member", ID: principal.PrincipalID}) {
		t.Fatalf("wrong actor %+v", response)
	}
	got = request(http.MethodDelete, path+"/2705", "", true)
	if got.Code != 200 {
		t.Fatalf("delete=%d %s", got.Code, got.Body)
	}
}
