package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Through the HTTP API, adding and then removing a reaction on another
// member's message leaves monolith with the removal.
func TestMessageReactionHTTPNotifiesLatestMemberState(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	actor := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root, err := store.AssignAIAgentTask(t.Context(), actor, "reaction-notify", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"})
	if err != nil {
		t.Fatal(err)
	}
	store.taskThreadMessages[root.ThreadID] = []AIAgentTaskThreadHistoryMessage{{MessageID: "asked", Role: AIAgentTaskThreadMessageRoleUser, AuthorPrincipalID: "user-2", Body: "please"}}
	recorder := &reactionNotificationRecorder{}
	server := NewServer(ServerConfig{AIAgentClient: store, Authorizer: reactionTestAuthorizer{actor}, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: recorder.server(t).URL, Key: "key"}})
	path := "/v2/client/workspaces/" + actor.WorkspaceID + "/ai-agent/tasks/" + root.TaskID + "/threads/" + root.ThreadID + "/messages/asked/reactions"
	for _, step := range []struct{ method, path, body string }{{http.MethodPost, path, `{"emoji":"1f44d"}`}, {http.MethodDelete, path + "/1f44d", ""}} {
		r := httptest.NewRequest(step.method, step.path, strings.NewReader(step.body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer test")
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("%s=%d %s", step.method, w.Code, w.Body)
		}
	}
	server.commentReactions.notificationWG.Wait()
	if got := recorder.sent(); len(got) == 0 || got[len(got)-1] != "" {
		t.Fatalf("sent %q, want the removal last", got)
	}
}
