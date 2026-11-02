package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type reactionNotificationRecorder struct {
	mu       sync.Mutex
	emojis   []string
	statuses []int
}

func (r *reactionNotificationRecorder) server(t *testing.T) *httptest.Server {
	t.Helper()
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/internal/agent-thread-reaction-notifications" {
			t.Errorf("unexpected path %s", req.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(req.Body).Decode(&body)
		r.mu.Lock()
		r.emojis = append(r.emojis, body["emoji"])
		status := http.StatusNoContent
		if len(r.statuses) > 0 {
			status, r.statuses = r.statuses[0], r.statuses[1:]
		}
		r.mu.Unlock()
		w.WriteHeader(status)
	}))
	t.Cleanup(endpoint.Close)
	return endpoint
}

func (r *reactionNotificationRecorder) sent() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.emojis...)
}

func reactionNotificationServer(t *testing.T, recorder *reactionNotificationRecorder) Server {
	t.Helper()
	return NewServer(ServerConfig{AIAgentClient: NewDevelopmentAIAgentClientStore(), CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: recorder.server(t).URL, Key: "key"}})
}

// An add whose handler reaches the queue after a later removal is dropped, so
// monolith keeps the removal.
func TestMemberReactionNotificationDropsOlderAddAfterRemoval(t *testing.T) {
	recorder := &reactionNotificationRecorder{}
	server := reactionNotificationServer(t, recorder)
	server.queueMemberReactionNotification(memberReactionNotification{key: "thread/message/user-2", revision: 2, body: map[string]string{"emoji": ""}})
	server.commentReactions.notificationWG.Wait()
	server.queueMemberReactionNotification(memberReactionNotification{key: "thread/message/user-2", revision: 1, body: map[string]string{"emoji": "1f44d"}})
	server.commentReactions.notificationWG.Wait()
	if got := recorder.sent(); len(got) != 1 || got[0] != "" {
		t.Fatalf("sent %q, want only the removal", got)
	}
}

func TestMemberReactionNotificationRetriesTransientFailures(t *testing.T) {
	previous := memberReactionNotificationRetryDelay
	memberReactionNotificationRetryDelay = time.Millisecond
	t.Cleanup(func() { memberReactionNotificationRetryDelay = previous })
	for _, tc := range []struct {
		name     string
		statuses []int
		want     int
	}{
		{name: "unavailable then delivered", statuses: []int{http.StatusServiceUnavailable}, want: 2},
		{name: "always unavailable", statuses: []int{503, 503, 503, 503, 503}, want: memberReactionNotificationMaxAttempts},
		{name: "not found is final", statuses: []int{http.StatusNotFound}, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &reactionNotificationRecorder{statuses: tc.statuses}
			server := reactionNotificationServer(t, recorder)
			server.queueMemberReactionNotification(memberReactionNotification{key: "thread/message/user-2", revision: 1, body: map[string]string{"emoji": "1f44d"}})
			server.commentReactions.notificationWG.Wait()
			if got := recorder.sent(); len(got) != tc.want {
				t.Fatalf("sent %d times, want %d", len(got), tc.want)
			}
		})
	}
}

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
