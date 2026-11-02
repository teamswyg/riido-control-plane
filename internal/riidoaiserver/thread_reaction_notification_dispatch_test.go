package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
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
